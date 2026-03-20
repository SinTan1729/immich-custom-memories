# Immich Custom Memories

This tool allows one to generate memories in [Immich](https://immich-app/immich) while filtering out certain faces, and tags.
I don't believe in deleting past images due to current feelings. (I've done it in the past, only regret it later.) But I don't
want them shoved in face through memories either. This tool aims to strike that balance.

# Installation

One can install it through AUR on Arch-based distros.

```sh
paru -S immich-custom-memories-bin

```

One can install it through [LURE](https://lure.sh) on pretty much any distro.

```sh
lure addrepo -n SinTan1729 -u https://github.com/SinTan1729/lure-repo
lure install immich-custom-memories

```

One can set up a memory generation job through cron or a systemd service. I use the following cronjob.

```cron
0 0 * * *       /usr/bin/immich-custom-memories

```

Make sure that the user has access to the config file. Also, make sure to enable the memories feature for the corresponding user
in Immich (Account Settings -> Features -> Time-based memories -> Enable), and disable the memory generation task (Administration ->
Nightly Task Settings -> Generate Memories -> Toggle off).

# Configuration

Configuration is read from `$XDG_CONFIG_HOME/immich-custom-memories/config.json`. An example config is provided in
[`example-config.json`](./example-config.json).

One can also pass a config file using `--config <file-name>`.

# Notes

- Immich v2.6.0 or higher is needed since we need immich-app/immich#26429 for the memories to properly appear on the timeline.
- Excluded people entry in the config supports both names and IDs of people.
