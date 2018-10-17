knife /kay-nife/ - a tool for knative yaml wrangling

# What dis?

It's a little tool for knative. Instead of being a whole CLI, it gives you some commands to spit out knative yaml which you can pipe to `"kubectl apply"`/`"kubetl create"` and friends. This is kind of nice, especially for development and learning about the system, 'cos you get to easily see what's happening under the covers and integrate with the rest of the yaml-ecosystem, but you don't need to write lots of yaml.

It's pronounced _kay-nife_. Sorry.

# Examples?

Sure. To create a build, you can do:

~~~~
knife generate-build -t buildpack -a arg1=foo -a arg2=bar -s github.com/foo/bar mybuild | kubectl apply -f -
~~~~

(Please remember to pronounce this "kay-nife" in your head).

To set up a source-to-service build you can do:

~~~~
knife generate-service myservice -s github.com/foo/bar -t buildpack -a IMAGE=docker.io/busybox docker.io/busybox echo hello | kubectl apply -f -
~~~~


Now let's do some routing!

~~~~
knife generate-route my-route -r revision1:100 -c configuration1:0:v2 | kubectl apply -f -
knife generate-route my-route -r revision1:80 -c configuration1:20:v2 | kubectl apply -f -
knife generate-route my-route -c configuration2:100 | kubectl apply -f -
~~~~

Similar stuff works for most other things.

*TIP*: For a diff showing what will change if you apply a generated object, you can pipe to `kubectl alpha diff -f - LAST LOCAL` instead of `kubectl apply -f -`.

# How about rapid local development?

Glad you asked! You can do a super-nice local-build-and-run-on-cluster for Go programs using the fantastic `ko apply` instead of `kubectl apply`:

~~~~
ko apply -L -f <( knife generate-service hello-world github.com/julz/knife/test/cmd/hello-world )
~~~~

NOTE: `ko apply -f` doesnt support `-` for stdin, so you can't just pipe to `ko apply -f -` :-(

What the above did is generate a Knative Service YML with 'github.com/julz/knife/test/cmd/hello-world/' as the Image, and then use `ko` to turn that in to a YML with a proper docker image URI and apply the manifest. The image gets built in your local minikube's docker (this also works fine for remote clusters, just lose the `-L` in the above command) so it's _blaaazing_ fast. See [the go-containerregistry repo](https://github.com/google/go-containerregistry/tree/master/cmd/ko) for more about `ko`.

Here's how to get a diff of what's about to be changed, you should see the image updated to the new built sha:

~~~~
ko resolve -f <(knife generate-service hello-world github.com/julz/knife/test/cmd/hello-world) | kubectl alpha diff -f - LAST LOCAL
~~~~

# What about Secrets and ServiceAccounts?

Sure!

~~~~
knife generate-secret the-secret -t git:github.com -t docker:docker.io | kubectl apply -f -
> Username: ...
> Password: ...

knife generate-service-account buildbot -s the-secret | kubectl apply -f -

knife generate-build mybuild --service-account buildbot -s github.com/julz/myapp -t kaniko | kubectl apply -f -
~~~~

# Anything else?

You can also use knife as a nice go library for building knative yml. e.g.

~~~~golang
func main() {
  json.NewEncoder(os.Stdout).Encode( 
    knative.NewBuild("my-build", knative.WithGitSource("github.com/foo/bar", "master"), knative.WithBuildTemplate("buildpack", knative.WithBuildTemplateArg("key", "value")
  )

  json.NewEncoder(os.Stdout).Encode( 
    knative.NewRunLatestService("my-service",
      knative.WithRevisionTemplate("docker.io/busybox", nil, nil), 
      knative.WithGitSource(knative.WithGitSource("github.com/foo/bar", "master"), 
      knative.WithBuildTemplate("buildpack"), 
      knative.WithBuildTemplateArg("key", "value"),
      knative.WithServiceAccount("buildbot"),
    ))
}
~~~~