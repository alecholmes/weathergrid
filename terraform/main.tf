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

# Policy Attachment on the role
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
                "s3:PutObject"
            ],
            "Resource": [
                "arn:aws:s3:::alecholmes-weatherdata/*"
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

# Create a lambda function
# In terraform ${path.module} is the current directory.
resource "aws_lambda_function" "syncweather" {
  filename         = data.archive_file.syncweather.output_path
  function_name    = "syncweather"
  role             = aws_iam_role.lambda_role.arn
  handler          = local.lambda_binary_filename
  source_code_hash = filebase64sha256(data.archive_file.syncweather.output_path)
  runtime          = "go1.x"
}

# Bucket to store weather data blobs
resource "aws_s3_bucket" "weatherdata" {
  bucket = "alecholmes-weatherdata"
  acl    = "private"
}