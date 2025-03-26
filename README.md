# scuttle
[![GoDoc](https://pkg.go.dev/badge/github.com/poseidon/scuttle.svg)](https://pkg.go.dev/github.com/poseidon/scuttle)
[![Quay](https://img.shields.io/badge/container-quay-green)](https://quay.io/repository/poseidon/scuttle)
[![Workflow](https://github.com/poseidon/scuttle/actions/workflows/build.yaml/badge.svg)](https://github.com/poseidon/scuttle/actions/workflows/build.yaml?query=branch%3Amain)
[![Sponsors](https://img.shields.io/github/sponsors/poseidon?logo=github)](https://github.com/sponsors/poseidon)
[![Mastodon](https://img.shields.io/badge/follow-news-6364ff?logo=mastodon)](https://fosstodon.org/@poseidon)

<img align="right" src="https://storage.googleapis.com/poseidon/scuttle.png" width="40%">

`scuttle` handles SIGTERM or spot termination notices by optionally draining or deleting the current Kubernetes node. It's best run as a systemd unit designed to gracefully stop on shutdown (shown below).

* Uncordon the node on start (reboot case)
* Handle SIGTERM (unit stop or shutdown)
* Monitor instance metadata for termination notices (AWS, Azure planned)
* Drain and/or delete (de-register) a Kubernetes node
* Evict Pods left by Kubelet's GracefulNodeShutdown

`scuttle` compliments Kubelet's [GracefulNodeShutdown](https://kubernetes.io/docs/concepts/cluster-administration/node-shutdown/#graceful-node-shutdown) feature, which only handles a part of gracefully stopping the Kubelet. See the [blog post](https://www.psdn.io/posts/kubelet-graceful-shutdown/) to learn more. In effect, `scuttle` is just a Go implementation of the bash scripts shown in the posts.

## Usage

`scuttle` must be run in a systemd unit that is designed to stop before shutdown. Systemd shutdown can be subtle. Read [systemd Shutdown Units](https://www.psdn.io/posts/systemd-shutdown-unit/) and [Kubelet Graceful Shutdown](https://www.psdn.io/posts/kubelet-graceful-shutdown/) for background.

```systemd
[Unit]
Description=Scuttle Kubelet before Shutdown
After=multi-user.target
[Service]
Type=simple
ExecStartPre=-/usr/bin/podman rm scuttle
ExecStart=/usr/bin/podman run \
  --name scuttle \
  --network host \
  --log-driver=k8s-file \
  --env KUBECONFIG=/var/lib/kubelet/kubeconfig \
  -v /var/lib/kubelet:/var/lib/kubelet:ro,z \
  --stop-timeout=60 \
  quay.io/poseidon/scuttle:v0.1.0 \
  -platform=aws
ExecStop=/usr/bin/podman stop scuttle
TimeoutStopSec=180
SuccessExitStatus=143
[Install]
WantedBy=multi-user.target
```

Users of Fedora CoreOS or Flatcar Linux should use the Butane Config:

```yaml
variant: fcos
version: 1.4.0
systemd:
  units:
    - name: scuttle.service
      contents: |
        [Unit]
        Description=Scuttle Kubelet before Shutdown
        After=multi-user.target
        [Service]
        Type=simple
        ExecStartPre=-/usr/bin/podman rm scuttle
        ExecStart=/usr/bin/podman run \
          --name scuttle \
          --network host \
          --log-driver=k8s-file \
          --env KUBECONFIG=/var/lib/kubelet/kubeconfig \
          -v /var/lib/kubelet:/var/lib/kubelet:ro,z \
          --stop-timeout=60 \
          quay.io/poseidon/scuttle:v0.1.0 \
          -platform=aws
        ExecStop=/usr/bin/podman stop scuttle
        TimeoutStopSec=180
        SuccessExitStatus=143
        [Install]
        WantedBy=multi-user.target
    - name: scuttle.path
      enabled: true
      contents: |
        [Unit]
        Description=Watch for Kubelet kubeconfig
        [Path]
        PathExists=/var/lib/kubelet/kubeconfig
        [Install]
        WantedBy=multi-user.target
```

### Configuration

Configure via flags.

| flag        | description  | default      |
|-------------|--------------|--------------|
| -node       | Kubernetes node name | $HOSTNAME |
| -platform   | Platform to poll for termination notices | none |
| -uncordon   | Uncordon node on start | true |
| -drain      | Drain node on stop     | true |
| -delete     | Delete node on stop    | true |
| -channel-id | Slack Channel ID       | ""   |
| -token      | Slack Bot Token        | ""   |
| -webhook    | Slack Webhook URL      | ""   |
| -log-level  | Logger level | info |
| -version    | Show version | NA   |
| -help       | Show help    | NA   |

Other values are set via environment variables.

| variable   | description            | default   |
|------------|------------------------|-----------|
| KUBECONFIG | Path to Kubeconfig     | ""        |
| HOSTNAME   | Current node name      | ""        |

### Spot Termination Notices

[AWS](https://aws.amazon.com/blogs/aws/new-ec2-spot-instance-termination-notices/) and [Azure](https://learn.microsoft.com/en-us/azure/virtual-machine-scale-sets/virtual-machine-scale-sets-terminate-notification) provide warnings via instance metadata (2 min) before spot terminations. `scuttle` can monitor platform specific instance metadata endpoints to trigger drain or delete actions before shutdown.

* [AWS Spot Termination Notifications](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/spot-instance-termination-notices.html)
* [Azure Spot Termination Notifications](https://learn.microsoft.com/en-us/azure/virtual-machine-scale-sets/virtual-machine-scale-sets-terminate-notification#get-terminate-notifications)

### Slack

`scuttle` integrates with Slack to post shutdown lifecycle events. Stay informed about spot terminations, shutdowns, drains, deletions.

* Create a [Slack App](https://api.slack.com/apps)

A token mode can post messages, thread replies, and add reactions.

* Grant bot token scopes `chat:write` and `reactions:write` to a channel via "OAuth & Permissions"
* Get the bot token
* Invite the app to a channel (`/invite @mybot`) and note the Channel ID

```
scuttle -channel-id C0FAKEFAKE -token bot-token
```

<img src="https://storage.googleapis.com/poseidon/scuttle-slack.png">

A webhook mode is available if you only have a Slack Webhook URL. However, it cannot thread replies or add reactions.

```
scuttle -webhook https://hooks.slack.com/...
```

## Development

To develop locally, build and run the executable.

### Static Binary

Build the static binary or container image.

```
make build
make image
```

### Run

Run the executable.

```
export KUBECONFIG=some-dev-kubeconfig
export HOSTNAME=node-name
./bin/scuttle
```

Use Ctrl-C to emulate a node shutdown.

```
INFO[0000] main: watch for interrupt signals
INFO[0000] main: starting scuttle
INFO[0000] start scuttle                                 hostname=ip-10-0-35-141
INFO[0000] scuttle: uncordon node                        hostname=ip-10-0-35-141
INFO[0000] drainer: uncordoning node                     node=ip-10-0-35-141
^C
INFO[0004] main: detected interrupt
INFO[0004] scuttle: stopping...                          hostname=ip-10-0-35-141
INFO[0004] scuttle: cordoning node                       hostname=ip-10-0-35-141
INFO[0004] drainer: cordoning node                       node=ip-10-0-35-141
INFO[0004] scuttle: draining node                        hostname=ip-10-0-35-141
INFO[0026] drainer: evicting pod                         node=ip-10-0-35-141 pod=redacted
INFO[0027] drainer: evicting pod                         node=ip-10-0-35-141 pod=redacted
...
INFO[0027] drainer: drained node                         node=ip-10-0-35-141
INFO[0027] scuttle: deleting node                        hostname=ip-10-0-35-141
INFO[0009] done
```
