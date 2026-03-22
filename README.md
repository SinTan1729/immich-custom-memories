# Immich Custom Memories

This tool lets you generate memories in [Immich](https://github.com/immich-app/immich) while filtering out certain faces and tags. I don’t
like deleting old photos just because of how I feel at a certain moment. I’ve done that before, only to regretted it later. But I also don’t
want those photos constantly showing up in memories. This tool is meant to strike a balance.

# Installation

You can install it through AUR on Arch-based distros:

```sh
paru -S immich-custom-memories-bin
```

Or through [LURE](https://lure.sh) on most distros:

```sh
lure addrepo -n SinTan1729 -u https://github.com/SinTan1729/lure-repo
lure install immich-custom-memories
```

You can also just clone the repo and run `make install`. (Use `make uninstall` to remove it.)

```sh
git clone https://github.com/SinTan1729/immich-custom-memories
cd immich-custom-memories
make install
```

# Post-install

It’s recommended to set up a memory generation job using cron or a systemd service. For example, here’s the cron job I use:

```cron
0 0 * * *       /usr/bin/immich-custom-memories
```

Make sure the user has access to the config file. Also ensure the memories feature is enabled for the user in Immich (Account Settings →
Features → Time-based memories → Enable), and disable Immich’s built-in memory generation task (Administration → Nightly Task Settings →
Generate Memories → Toggle off).

# Configuration

Configuration is read from `$XDG_CONFIG_HOME/immich-custom-memories/config.json`. An example config is available in
[`example-config.json`](./example-config.json).

You can also pass a config file using `--config <file-name>`. Using a local/internal URI is recommended for better performance.

Make sure that the API key used has at least the following permissions.

```
[ memory.create, memory.read, memory.delete, asset.read ]
```

# Notes

- Requires Immich v2.6.0 or higher, since it depends on immich-app/immich#26429 for memories to show up properly in the timeline.
- The `excludedPeople` field in the config supports both names and IDs.
