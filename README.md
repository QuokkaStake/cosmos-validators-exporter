# cosmos-validators-exporter

![Latest release](https://img.shields.io/github/v/release/freak12techno/cosmos-validators-exporter)
[![Actions Status](https://github.com/freak12techno/cosmos-validators-exporter/workflows/test/badge.svg)](https://github.com/freak12techno/cosmos-validators-exporter/actions)

cosmos-validators-exporter is a Prometheus scraper that fetches validators stats from an LCD server exposed by a fullnode.

## What can I use it for?

You can use it for building a validator(s) dashboard(s) to visualize your validators metrics (like total staked amount, delegators count, etc.) as well as building alerts for your validators' metrics (like validator status/ranking/total staked amount/etc.)

Here's an example of the dashboard we're using in production:

![Validators dashboard](https://raw.githubusercontent.com/freak12techno/cosmos-validators-exporter/master/images/01.jpg)

## How can I set it up?

First of all, you need to download the latest release from [the releases page](https://github.com/freak12techno/cosmos-validators-exporter/releases/). After that, you should unzip it and you are ready to go:

```sh
wget <the link from the releases page>
tar xvfz <filename you just downloaded>
./cosmos-validators-exporter
```

To run it detached, you need to run it as a systemd service. First of all, we have to copy the file to the system apps folder:

```sh
sudo cp ./cosmos-validators-exporter /usr/bin
```

Then we need to create a systemd service for our app:

```sh
sudo nano /etc/systemd/system/cosmos-validators-exporter.service
```

You can use this template (change the user to whatever user you want this to be executed from. It's advised to create a separate user for that instead of running it from root):

```
[Unit]
Description=Cosmos Validators Exporter
After=network-online.target

[Service]
User=<username>
TimeoutStartSec=0
CPUWeight=95
IOWeight=95
ExecStart=cosmos-validators-exporter --config <path to config>
Restart=always
RestartSec=2
LimitNOFILE=800000
KillSignal=SIGTERM

[Install]
WantedBy=multi-user.target
```

Then we'll add this service to the autostart and run it:

```sh
sudo systemctl daemon-reload # reflect the systemd file change
sudo systemctl enable cosmos-validators-exporter # enable the scraper to run on system startup
sudo systemctl start cosmos-validators-exporter # start it
sudo systemctl status cosmos-validators-exporter # validate it's running
```

If you need to, you can also see the logs of the process:

```sh
sudo journalctl -u cosmos-validators-exporter -f --output cat
```

## How can I scrape data from it?

Here's the example of the Prometheus config you can use for scraping data:

```yaml
scrape-configs:
  - job_name:       'cosmos-validators-exporter'
    scrape_interval: 30s
    static_configs:
      - targets:
        - localhost:9560 # replace localhost with scraper IP if it's on the other host
```

Then restart Prometheus and you're good to go!

All of the metrics provided by cosmos-validators-exporter have the `cosmos_validators_exporter_` as a prefix. For the full list of metrics, try running `curl localhost:9560/metrics` (or your host/port, if it's non-standard) and see the list of metrics there.

## How can I configure it?

All of the configuration is done via the .toml config file, which is passed to the application via the `--config` app parameter. Check `config.example.toml` for a config reference.

## How can I contribute?

Bug reports and feature requests are always welcome! If you want to contribute, feel free to open issues or PRs.
