package recipe_test

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	bap "code.cloudfoundry.org/buildpackapplifecycle"
	. "code.cloudfoundry.org/eirini/recipe"
	"code.cloudfoundry.org/eirini/recipe/recipefakes"
)

var _ = Describe("PacksExecutor", func() {

	var (
		executor       Executor
		commander      *recipefakes.FakeCommander
		resultModifier *recipefakes.FakeStagingResultModifier
		tmpfile        *os.File
		resultContents string
	)

	createTmpFile := func() {
		var err error

		tmpfile, err = ioutil.TempFile("", "metadata_result")
		Expect(err).ToNot(HaveOccurred())

		_, err = tmpfile.Write([]byte(resultContents))
		Expect(err).ToNot(HaveOccurred())

		err = tmpfile.Close()
		Expect(err).ToNot(HaveOccurred())
	}

	BeforeEach(func() {
		commander = new(recipefakes.FakeCommander)
		resultModifier = new(recipefakes.FakeStagingResultModifier)

		resultModifier.ModifyStub = func(result bap.StagingResult) (bap.StagingResult, error) {
			return result, nil
		}

		resultContents = `{"lifecycle_type":"no-type", "execution_metadata":"data"}`
	})

	JustBeforeEach(func() {
		createTmpFile()
		packsConf := PacksBuilderConf{
			BuildpacksDir:             "/var/lib/buildpacks",
			OutputDropletLocation:     "/out/droplet.tgz",
			OutputBuildArtifactsCache: "/cache/cache.tgz",
			OutputMetadataLocation:    tmpfile.Name(),
		}

		executor = &PacksExecutor{
			Conf:      packsConf,
			Commander: commander,
		}

	})

	AfterEach(func() {
		Expect(os.Remove(tmpfile.Name())).To(Succeed())
	})

	Context("When executing a recipe", func() {

		var (
			err    error
			server *ghttp.Server
		)

		BeforeEach(func() {
			server = ghttp.NewServer()
			server.RouteToHandler("PUT", "/stage/staging-guid/completed",
				ghttp.VerifyJSON(`{
						"task_guid": "staging-guid",
						"failed": false,
						"failure_reason": "",
						"result": "{\"lifecycle_metadata\":{\"detected_buildpack\":\"\",\"buildpacks\":null},\"process_types\":null,\"execution_metadata\":\"data\",\"lifecycle_type\":\"no-type\"}",
						"annotation": "{\"lifecycle\":\"\",\"completion_callback\":\"completion-call-me-back\"}",
						"created_at": 0
					}`),
			)
		})

		JustBeforeEach(func() {
			recipeConf := Config{
				AppID:              "app-id",
				StagingGUID:        "staging-guid",
				CompletionCallback: "completion-call-me-back",
				EiriniAddr:         server.URL(),
				DropletUploadURL:   "droplet.eu/upload",
				PackageDownloadURL: server.URL() + "app-id",
			}
			err = executor.ExecuteRecipe(recipeConf)
		})

		AfterEach(func() {
			server.Close()
		})

		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("should run the packs builder", func() {
			Expect(commander.ExecCallCount()).To(Equal(1))

			cmd, args := commander.ExecArgsForCall(0)
			Expect(cmd).To(Equal("/packs/builder"))
			Expect(args).To(ConsistOf(
				"-buildpacksDir", "/var/lib/buildpacks",
				"-outputDroplet", "/out/droplet.tgz",
				"-outputBuildArtifactsCache", "/cache/cache.tgz",
				"-outputMetadata", tmpfile.Name(),
			))
		})

		// Context("and download or extract of app bits fails", func() {
		//
		// 	BeforeEach(func() {
		// 		installer.InstallReturns(errors.New("boom"))
		// 		server.RouteToHandler("PUT", "/stage/staging-guid/completed",
		// 			ghttp.VerifyJSON(`{
		// 				"task_guid": "staging-guid",
		// 				"failed": true,
		// 				"failure_reason": "boom",
		// 				"result": "",
		// 				"annotation": "{\"lifecycle\":\"\",\"completion_callback\":\"completion-call-me-back\"}",
		// 				"created_at": 0
		// 			}`),
		// 		)
		// 	})
		//
		// 	It("should return an error", func() {
		// 		Expect(err).To(HaveOccurred())
		// 	})
		//
		// 	It("should send completion response with a failure", func() {
		// 		Expect(server.ReceivedRequests()).To(HaveLen(1))
		// 	})
		//
		// })

		// Context("and it fails to execute packs builder", func() {
		//
		// 	BeforeEach(func() {
		// 		commander.ExecReturns(errors.New("boomz"))
		// 		server.RouteToHandler("PUT", "/stage/staging-guid/completed",
		// 			ghttp.VerifyJSON(`{
		// 				"task_guid": "staging-guid",
		// 				"failed": true,
		// 				"failure_reason": "boomz",
		// 				"result": "",
		// 				"annotation": "{\"lifecycle\":\"\",\"completion_callback\":\"completion-call-me-back\"}",
		// 				"created_at": 0
		// 			}`),
		// 		)
		// 	})
		//
		// 	It("should return an error", func() {
		// 		Expect(err).To(HaveOccurred())
		// 	})
		//
		// 	It("should send completion response with a failure", func() {
		// 		Expect(server.ReceivedRequests()).To(HaveLen(1))
		// 	})
		// })

		// Context("and it fails to upload the droplet", func() {
		//
		// 	BeforeEach(func() {
		// 		uploader.UploadReturns(errors.New("booma"))
		// 		server.RouteToHandler("PUT", "/stage/staging-guid/completed",
		// 			ghttp.VerifyJSON(`{
		// 				"task_guid": "staging-guid",
		// 				"failed": true,
		// 				"failure_reason": "booma",
		// 				"result": "",
		// 				"annotation": "{\"lifecycle\":\"\",\"completion_callback\":\"completion-call-me-back\"}",
		// 				"created_at": 0
		// 			}`),
		// 		)
		// 	})
		//
		// 	It("should return an error", func() {
		// 		Expect(err).To(HaveOccurred())
		// 	})
		//
		// 	It("should send completion response with a failure", func() {
		// 		Expect(server.ReceivedRequests()).To(HaveLen(1))
		// 	})
		// })

		// Context("and eirini returns response with failure status", func() {
		//
		// 	BeforeEach(func() {
		// 		server.RouteToHandler("PUT", "/stage/staging-guid/completed",
		// 			ghttp.RespondWith(http.StatusInternalServerError, ""),
		// 		)
		// 	})
		//
		// 	It("should return an error", func() {
		// 		Expect(server.ReceivedRequests()).To(HaveLen(1))
		// 		Expect(err).To(HaveOccurred())
		// 	})
		//
		// })

	})

})
