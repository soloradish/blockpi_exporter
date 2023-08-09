# BlockPi Exporter
Prometheus exporter for BlockPi metrics, written in Go.

## Overview
The blockpi_exporter is a service that fetches data from the BlockPi API and exports them as Prometheus metrics. It's an ideal tool for those looking to monitor their BlockPi metrics through Prometheus.

## Getting Started
Prerequisites
* Go
* Prometheus

## Installation
Clone the repository:
```bash
git clone https://github.com/soloradish/blockpi_exporter.git
```
Navigate to the project directory and build:
```bash
cd blockpi_exporter
go build .
```

## Configuration with Dotenv
This project uses dotenv for easy configuration management. To set up your configuration:

Create a .env file in the root directory of the project.

Add your config to the .env file:
```
# .env
BLOCKPI_API_KEY=<your_blockpi_api_key> # required
BLOCKPI_LISTEN_PORT=<port_you_want_to_use> # optional, defaults to 8080
```

The blockpi_exporter will automatically load the API key from the .env file when it starts up. Ensure you don't commit your .env file to any public repositories to keep your API key secure.

You can use the .env file or set environment variables to configure blockpi_exporter like this:

```bash
BLOCKPI_API_KEY=<your_blockpi_api_key> ./blockpi_exporter
```

## Configure your Prometheus instance to scrape this exporter:
```yaml
scrape_configs:
- job_name: 'blockpi_exporter'
  static_configs:
    - targets: ['localhost:8080'] # or whatever port you set in your .env file
```
## Contributing
Contributions are welcome! Please create an issue or submit a pull request for enhancements, bug fixes, or features.

## License
This project is licensed under the MIT License. Check the LICENSE file for more details.