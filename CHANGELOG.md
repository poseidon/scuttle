# scuttle

Notable changes between versions.

## Latest

* Add Slack threading and reactions ([#7](https://github.com/poseidon/scuttle/pull/7))
  * Add `-channel-id` flag to set a channel
  * Add `-token` flag to set a Slack token
* Add Slack notifications for lifecycle events ([#6](https://github.com/poseidon/scuttle/pull/6))
  * Add `-webhook` flag for basic Slack notifications

## v0.1.0

* Initial port from bash script to Go
* Make uncordon, drain, and delete optional
* Poll AWS spot instance termination notices
* Drop requirement that `kubectl` be present
