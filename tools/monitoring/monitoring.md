# Setting up a Monitoring Dashboard

Simply running `docker compose up` in the `/wasp/tools/monitoring` directory will

- Bring up a wasp node with the metrics plugin turned on and prometheus metrics exposed at `:2112`. 
- A prometheus server running on port `9091` configured to scrape metrics from the wasp node at port `2112`
- Grafana using prometheus at `9091` as default data source.

You can customise the configuration to your liking by editing the respective files.

### Note (TODO)

1) No dashboards configured yet 
