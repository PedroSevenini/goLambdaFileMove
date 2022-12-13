provider "aws" {
  region = "${var.region}"
  profile = "dev-terraform"
}

locals {
  lambda_handler  = "goLambdaFileMove"
  name            = "golambda-FileMove"
}

data "archive_file" "lambda_zip" {
  type        = "zip"
  source_file = "../bin/goLambdaFileMove"
  output_path = "bin/goLambdaFileMove.zip"
}

resource "aws_lambda_function" "golambda-FileMove" {
  function_name     = local.name
  handler           = local.lambda_handler
  runtime           = "go1.x"
  role              = "${var.lambda_role_arn}"
  filename          = data.archive_file.lambda_zip.output_path
  source_code_hash = filebase64sha256(data.archive_file.lambda_zip.output_path)
  memory_size       = 128
  timeout           = 30
  environment {
    variables = {
      SNS_ARN = aws_sns_topic.sns-FileMove.arn
    }
  }
}

resource "aws_s3_bucket" "bucket" {
  bucket = var.bucket_name
  tags = {
    Environment = var.environment
  }
}

resource "aws_s3_bucket_notification" "aws-lambda-trigger" {
  bucket = aws_s3_bucket.bucket.id
  lambda_function {
    lambda_function_arn = aws_lambda_function.golambda-FileMove.arn
    events              = ["s3:ObjectCreated:*"]

  }
}

resource "aws_lambda_permission" "fileMove-bucket-permission" {
  statement_id  = "AllowS3Invoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.golambda-FileMove.function_name
  principal     = "s3.amazonaws.com"
  source_arn    = "arn:aws:s3:::${aws_s3_bucket.bucket.id}"
}

resource "aws_sns_topic" "sns-FileMove" {
  name = "sns-FileMove"
}