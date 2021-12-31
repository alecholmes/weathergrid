# Weather

Simple webapp displaying a grid of weather forecasts. A deployment of this using Bay Area locales is running
at [https://alecholmes.com/weather](https://alecholmes.com/weather).

The frontend is implemented completely in [index.html](site/index.html) without using any third party dependencies.

Weather data is stored in a public JSON blob which is periodically updated by the backend.

The backend is a Golang CLI app that will run as an AWS lambda if no arguments are provided.
See [main.tf](terraform/main.tf) for all the necessary boilerplate to deploy the backend.

## Building and deploying the lambda

Prerequisites:

* [OpenWeather](https://openweathermap.org/api) API key
* AWS account
* Golang toolchain
* Terraform

```shell
make build-linux

cd terraform
terraform apply
```

Terraform will need to run in the context of an AWS user with the following permission policies (these can be scoped
down):

* IAMFullAccess
* AmazonS3FullAccess
* CloudWatchEventsFullAccess
* AWSLambda_FullAccess

## Creating weather snapshot JSON locally

```shell
# If -snapshot-to-s3 is omitted, the snapshot is written to stdout instead.
go run ./cmd/syncweather -local -config config/config.json -snapshot-to-s3
```

# Notes

The OpenWeather API free tier only allows 1000 API calls per day. Since there is an API call for each location, the
number of locations and lambda frequency must be tweaked to remain within this limit. If data is updated hourly, there
can be no more than 41 locations.
