package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"code.cloudfoundry.org/eirini"
	cmdcommons "code.cloudfoundry.org/eirini/cmd"
	"code.cloudfoundry.org/eirini/events"
	"code.cloudfoundry.org/eirini/handler"
	k8sevent "code.cloudfoundry.org/eirini/k8s/informers/event"
	k8sroute "code.cloudfoundry.org/eirini/k8s/informers/route"
	"code.cloudfoundry.org/eirini/metrics"
	"code.cloudfoundry.org/eirini/route"
	"code.cloudfoundry.org/eirini/stager"
	"code.cloudfoundry.org/eirini/util"
	loggregator "code.cloudfoundry.org/go-loggregator"
	"code.cloudfoundry.org/lager"

	yaml "gopkg.in/yaml.v2"

	"code.cloudfoundry.org/eirini/bifrost"
	"code.cloudfoundry.org/eirini/k8s"
	"code.cloudfoundry.org/tps/cc_client"
	nats "github.com/nats-io/go-nats"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	// For gcp and oidc authentication
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "connects CloudFoundry with Kubernetes",
	Run:   connect,
}

func connect(cmd *cobra.Command, args []string) {
	path, err := cmd.Flags().GetString("config")
	cmdcommons.ExitWithError(err)

	if path == "" {
		cmdcommons.ExitWithError(errors.New("--config is missing"))
	}

	cfg := setConfigFromFile(path)
	stager := initStager(cfg)
	bifrost := initBifrost(cfg)

	launchRouteEmitter(
		cfg.Properties.KubeConfigPath,
		cfg.Properties.KubeNamespace,
		cfg.Properties.NatsPassword,
		cfg.Properties.NatsIP,
	)

	tlsConfig, err := loggregator.NewIngressTLSConfig(
		cfg.Properties.LoggregatorCAPath,
		cfg.Properties.LoggregatorCertPath,
		cfg.Properties.LoggregatorKeyPath,
	)
	cmdcommons.ExitWithError(err)

	loggregatorClient, err := loggregator.NewIngressClient(
		tlsConfig,
		loggregator.WithAddr(cfg.Properties.LoggregatorAddress),
	)
	cmdcommons.ExitWithError(err)
	defer func() {
		if err = loggregatorClient.CloseSend(); err != nil {
			cmdcommons.ExitWithError(err)
		}
	}()
	launchMetricsEmitter(
		cfg.Properties.KubeConfigPath,
		fmt.Sprintf("%s/namespaces/%s/pods", cfg.Properties.MetricsSourceAddress, cfg.Properties.KubeNamespace),
		loggregatorClient,
		cfg.Properties.KubeNamespace,
	)

	launchEventReporter(
		cfg.Properties.KubeConfigPath,
		cfg.Properties.CcInternalAPI,
		cfg.Properties.CCCAPath,
		cfg.Properties.CCCertPath,
		cfg.Properties.CCKeyPath,
		cfg.Properties.KubeNamespace,
	)

	handlerLogger := lager.NewLogger("handler")
	handlerLogger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))
	handler := handler.New(bifrost, stager, handlerLogger)

	log.Println("opi connected")
	log.Fatal(http.ListenAndServe("0.0.0.0:8085", handler))
}

func initStager(cfg *eirini.Config) eirini.Stager {
	clientset := cmdcommons.CreateKubeClient(cfg.Properties.KubeConfigPath)
	taskDesirer := &k8s.TaskDesirer{
		Namespace:       cfg.Properties.KubeNamespace,
		CCUploaderIP:    cfg.Properties.CcUploaderIP,
		CertsSecretName: cfg.Properties.CCCertsSecretName,
		Client:          clientset,
	}

	stagerCfg := eirini.StagerConfig{
		EiriniAddress: cfg.Properties.EiriniAddress,
		Image:         getStagerImage(cfg),
	}

	httpClient, err := util.CreateTLSHTTPClient(
		[]util.CertPaths{
			{
				Crt: cfg.Properties.CCCertPath,
				Key: cfg.Properties.CCKeyPath,
				Ca:  cfg.Properties.CCCAPath,
			},
		},
	)
	if err != nil {
		panic(errors.Wrap(err, "failed to create stager http client"))
	}

	return stager.New(taskDesirer, httpClient, stagerCfg)
}

func initBifrost(cfg *eirini.Config) eirini.Bifrost {
	syncLogger := lager.NewLogger("bifrost")
	syncLogger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))
	kubeNamespace := cfg.Properties.KubeNamespace
	clientset := cmdcommons.CreateKubeClient(cfg.Properties.KubeConfigPath)
	desirer := k8s.NewStatefulSetDesirer(clientset, kubeNamespace)
	convertLogger := lager.NewLogger("convert")
	convertLogger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))
	registryIP := cfg.Properties.RegistryAddress
	converter := bifrost.NewConverter(&util.TruncatedSHA256Hasher{}, convertLogger, registryIP)

	return &bifrost.Bifrost{
		Converter: converter,
		Desirer:   desirer,
		Logger:    syncLogger,
	}
}

func setConfigFromFile(path string) *eirini.Config {
	fileBytes, err := ioutil.ReadFile(filepath.Clean(path))
	cmdcommons.ExitWithError(err)

	var Conf eirini.Config
	err = yaml.Unmarshal(fileBytes, &Conf)
	cmdcommons.ExitWithError(err)

	return &Conf
}

func initConnect() {
	connectCmd.Flags().StringP("config", "c", "", "Path to the erini config file")
}

func launchRouteEmitter(kubeConfigPath, namespace, natsPassword, natsIP string) {
	nc, err := nats.Connect(util.GenerateNatsURL(natsPassword, natsIP))
	cmdcommons.ExitWithError(err)

	clientset := cmdcommons.CreateKubeClient(kubeConfigPath)
	syncPeriod := 10 * time.Second
	workChan := make(chan *route.Message)
	instanceInformer := k8sroute.NewInstanceChangeInformer(clientset, syncPeriod, namespace)
	uriInformer := k8sroute.NewURIChangeInformer(clientset, syncPeriod, namespace)
	re := route.NewEmitter(&route.NATSPublisher{NatsClient: nc}, workChan, &route.SimpleLoopScheduler{}, os.Stdout)

	go re.Start()
	go instanceInformer.Start(workChan)
	go uriInformer.Start(workChan)
}

func launchMetricsEmitter(kubeConfigPath, source string, loggregatorClient *loggregator.IngressClient, namespace string) {
	work := make(chan []metrics.Message, 20)
	clientset := cmdcommons.CreateKubeClient(kubeConfigPath)
	podClient := clientset.CoreV1().Pods(namespace)
	collector := k8s.NewMetricsCollector(work, &route.SimpleLoopScheduler{}, source, podClient)
	forwarder := metrics.NewLoggregatorForwarder(loggregatorClient)
	emitter := metrics.NewEmitter(work, &route.SimpleLoopScheduler{}, forwarder)

	go collector.Start()
	go emitter.Start()
}

func launchEventReporter(kubeConfigPath, uri, ca, cert, key, namespace string) {
	work := make(chan events.CrashReport, 20)
	tlsConf, err := cc_client.NewTLSConfig(cert, key, ca)
	cmdcommons.ExitWithError(err)

	client := cc_client.NewCcClient(uri, tlsConf)
	reporter := events.NewCrashReporter(work, &route.SimpleLoopScheduler{}, client, lager.NewLogger("instance-crash-reporter"))
	clientset := cmdcommons.CreateKubeClient(kubeConfigPath)
	crashInformer := k8sevent.NewCrashInformer(clientset, 0, namespace, work, make(chan struct{}))

	go crashInformer.Start()
	go reporter.Run()
}

func getStagerImage(cfg *eirini.Config) string {
	return cfg.Properties.StagerImage
}
