terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.0"
    }
  }
}

provider "aws" {
  region = "us-east-1"
}

locals {
  s3_prefix = "alecholmes"

  lambda_binary_filename = "syncweather_linux_amd64"
}

resource "aws_iam_role" "lambda_role" {
  name               = "weathersync_lambda_role"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

# IAM policy for logging from a lambda
resource "aws_iam_policy" "lambda_logging" {
  name        = "weathersync_lambda_logging_policy"
  path        = "/"
  description = "IAM policy for logging from a lambda"
  policy      = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:*:*:*",
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "lambda_logging_attach" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = aws_iam_policy.lambda_logging.arn
}

resource "aws_iam_policy" "lambda_s3" {
  name        = "weathersync_lambda_s3_policy"
  path        = "/"
  description = "IAM policy for S3 bucket puts from a lambda"
  policy      = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:GetObject",
                "s3:PutObject"
            ],
            "Resource": [
                "arn:aws:s3:::${aws_s3_bucket.weatherdata.bucket}/*"
            ]
        },
        {
            "Effect": "Allow",
            "Action": [
                "s3:GetObject"
            ],
            "Resource": [
                "arn:aws:s3:::${aws_s3_bucket.internal.bucket}/*"
            ]
        }
    ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "lambda_s3_attach" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = aws_iam_policy.lambda_s3.arn
}

# Generates an archive from content, a file, or a directory of files
data "archive_file" "syncweather" {
  type        = "zip"
  source_file = "${path.module}/../bin/${local.lambda_binary_filename}"
  output_path = "bin/syncweather.zip"
}

resource "aws_lambda_function" "syncweather" {
  filename         = data.archive_file.syncweather.output_path
  function_name    = "syncweather"
  role             = aws_iam_role.lambda_role.arn
  handler          = local.lambda_binary_filename
  source_code_hash = filebase64sha256(data.archive_file.syncweather.output_path)
  runtime          = "go1.x"
  timeout          = 15
}

# Bucket to store public weather data blobs
resource "aws_s3_bucket" "weatherdata" {
  bucket = "${local.s3_prefix}-weatherdata-snapshots"
  acl    = "private"

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET"]
    allowed_origins = ["*"]
    expose_headers  = ["ETag"]
    max_age_seconds = 300
  }
}

# Make weather snapshots in weatherdata bucket publicly readable.
resource "aws_s3_bucket_policy" "weatherdata" {
  bucket = aws_s3_bucket.weatherdata.bucket
  policy = <<EOF
{
  "Version":"2012-10-17",
  "Statement":[
    {
      "Effect":"Allow",
      "Principal": "*",
      "Action":["s3:GetObject"],
      "Resource":["arn:aws:s3:::${aws_s3_bucket.weatherdata.bucket}/snapshot_*"]
      }
  ]
}
EOF
}

# Bucket to store private lambda weather config.
resource "aws_s3_bucket" "internal" {
  bucket = "${local.s3_prefix}-weatherdata-internal"
  acl    = "private"
}

resource "aws_s3_bucket_object" "config" {
  bucket = aws_s3_bucket.internal.bucket
  key    = "syncweather-config.json"
  source = "../config/config.json"
  etag   = filemd5("../config/config.json")
}

# Schedule to run the lambda
resource "aws_cloudwatch_event_rule" "syncweather_schedule" {
  name                = "syncweather-schedule"
  schedule_expression = "cron(30 * * * ? *)"
}

resource "aws_cloudwatch_event_target" "trigger_syncweather" {
  rule      = aws_cloudwatch_event_rule.syncweather_schedule.name
  target_id = "lambda"
  arn       = aws_lambda_function.syncweather.arn
  input     = <<EOF
{
  "config_bucket_name": "${aws_s3_bucket.internal.bucket}",
  "config_object_name": "${aws_s3_bucket_object.config.key}"
}
EOF
}

resource "aws_lambda_permission" "allow_cloudwatch_to_call_syncweather" {
  statement_id  = "AllowExecutionFromCloudWatch"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.syncweather.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.syncweather_schedule.arn
}
