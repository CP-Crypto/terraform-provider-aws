package logs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccLogsSubscriptionFilter_basic(t *testing.T) {
	var filter cloudwatchlogs.SubscriptionFilter
	lambdaFunctionResourceName := "aws_lambda_function.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"
	resourceName := "aws_cloudwatch_log_subscription_filter.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriptionFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriptionFilterConfig_destinationARNLambda(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriptionFilterExists(resourceName, &filter),
					resource.TestCheckResourceAttrPair(resourceName, "destination_arn", lambdaFunctionResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "distribution", "ByLogStream"),
					resource.TestCheckResourceAttr(resourceName, "filter_pattern", "logtype test"),
					resource.TestCheckResourceAttrPair(resourceName, "log_group_name", logGroupResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSubscriptionFilterImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLogsSubscriptionFilter_many(t *testing.T) {
	resourceName := "aws_cloudwatch_log_subscription_filter.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriptionFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriptionFilterConfig_destinationARNLambdaMany(rName, 2), // This is the default limit of subscription filters on an account
				Check:  testAccCheckSubscriptionFilterManyExists(resourceName, 2),
			},
		},
	})
}

func TestAccLogsSubscriptionFilter_disappears(t *testing.T) {
	var filter cloudwatchlogs.SubscriptionFilter
	resourceName := "aws_cloudwatch_log_subscription_filter.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriptionFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriptionFilterConfig_destinationARNLambda(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriptionFilterExists(resourceName, &filter),
					acctest.CheckResourceDisappears(acctest.Provider, tflogs.ResourceSubscriptionFilter(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLogsSubscriptionFilter_Disappears_logGroup(t *testing.T) {
	var filter cloudwatchlogs.SubscriptionFilter
	var logGroup cloudwatchlogs.LogGroup
	logGroupResourceName := "aws_cloudwatch_log_group.test"
	resourceName := "aws_cloudwatch_log_subscription_filter.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriptionFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriptionFilterConfig_destinationARNLambda(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriptionFilterExists(resourceName, &filter),
					testAccCheckGroupExists(logGroupResourceName, &logGroup),
					acctest.CheckResourceDisappears(acctest.Provider, tflogs.ResourceGroup(), logGroupResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLogsSubscriptionFilter_DestinationARN_kinesisDataFirehose(t *testing.T) {
	var filter cloudwatchlogs.SubscriptionFilter
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	resourceName := "aws_cloudwatch_log_subscription_filter.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriptionFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriptionFilterConfig_destinationARNKinesisDataFirehose(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriptionFilterExists(resourceName, &filter),
					resource.TestCheckResourceAttrPair(resourceName, "destination_arn", firehoseResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSubscriptionFilterImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLogsSubscriptionFilter_DestinationARN_kinesisStream(t *testing.T) {
	var filter cloudwatchlogs.SubscriptionFilter
	kinesisStream := "aws_kinesis_stream.test"
	resourceName := "aws_cloudwatch_log_subscription_filter.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriptionFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriptionFilterConfig_destinationARNKinesisStream(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriptionFilterExists(resourceName, &filter),
					resource.TestCheckResourceAttrPair(resourceName, "destination_arn", kinesisStream, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSubscriptionFilterImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLogsSubscriptionFilter_distribution(t *testing.T) {
	var filter cloudwatchlogs.SubscriptionFilter
	resourceName := "aws_cloudwatch_log_subscription_filter.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriptionFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriptionFilterConfig_distribution(rName, "Random"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriptionFilterExists(resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, "distribution", "Random"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSubscriptionFilterImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccSubscriptionFilterConfig_distribution(rName, "ByLogStream"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriptionFilterExists(resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, "distribution", "ByLogStream"),
				),
			},
		},
	})
}

func TestAccLogsSubscriptionFilter_roleARN(t *testing.T) {
	var filter cloudwatchlogs.SubscriptionFilter
	iamRoleResourceName1 := "aws_iam_role.test"
	iamRoleResourceName2 := "aws_iam_role.test2"
	resourceName := "aws_cloudwatch_log_subscription_filter.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriptionFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriptionFilterConfig_roleARN1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriptionFilterExists(resourceName, &filter),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", iamRoleResourceName1, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSubscriptionFilterImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccSubscriptionFilterConfig_roleARN2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriptionFilterExists(resourceName, &filter),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", iamRoleResourceName2, "arn"),
				),
			},
		},
	})
}

func testAccCheckSubscriptionFilterDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LogsConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_log_subscription_filter" {
			continue
		}

		_, err := tflogs.FindSubscriptionFilterByTwoPartKey(context.Background(), conn, rs.Primary.Attributes["log_group_name"], rs.Primary.Attributes["name"])

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("CloudWatch Logs Filter Subscription still exists: %s", rs.Primary.ID)
	}

	return nil
}

func testAccCheckSubscriptionFilterExists(n string, v *cloudwatchlogs.SubscriptionFilter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CloudWatch Logs Filter Subscription ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LogsConn

		output, err := tflogs.FindSubscriptionFilterByTwoPartKey(context.Background(), conn, rs.Primary.Attributes["log_group_name"], rs.Primary.Attributes["name"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccSubscriptionFilterImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		logGroupName := rs.Primary.Attributes["log_group_name"]
		filterNamePrefix := rs.Primary.Attributes["name"]
		stateID := fmt.Sprintf("%s|%s", logGroupName, filterNamePrefix)

		return stateID, nil
	}
}

func testAccCheckSubscriptionFilterManyExists(basename string, n int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for i := 0; i < n; i++ {
			n := fmt.Sprintf("%s.%d", basename, i)
			var v cloudwatchlogs.SubscriptionFilter

			err := testAccCheckSubscriptionFilterExists(n, &v)(s)

			if err != nil {
				return err
			}
		}

		return nil
	}
}

func testAccSubscriptionFilterConfig_kinesisDataFirehoseBase(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_cloudwatch_log_group" "test" {
  name              = %[1]q
  retention_in_days = 1
}

resource "aws_iam_role" "firehose" {
  name = "%[1]s-firehose"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "${data.aws_caller_identity.current.account_id}"
        }
      }
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "firehose" {
  role = aws_iam_role.firehose.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": [
        "s3:AbortMultipartUpload",
        "s3:GetBucketLocation",
        "s3:GetObject",
        "s3:ListBucket",
        "s3:ListBucketMultipartUploads",
        "s3:PutObject"
      ],
      "Resource": [
        "${aws_s3_bucket.test.arn}",
        "${aws_s3_bucket.test.arn}/*"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_role" "cloudwatchlogs" {
  name = "%[1]s-cloudwatchlogs"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "logs.${data.aws_region.current.name}.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "cloudwatchlogs" {
  role = aws_iam_role.cloudwatchlogs.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "firehose:*",
      "Resource": "arn:${data.aws_partition.current.partition}:firehose:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"
    },
    {
      "Effect": "Allow",
      "Action": "iam:PassRole",
      "Resource": "${aws_iam_role.cloudwatchlogs.arn}"
    }
  ]
}
EOF
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  destination = "extended_s3"
  name        = %[1]q

  extended_s3_configuration {
    bucket_arn = aws_s3_bucket.test.arn
    role_arn   = aws_iam_role.firehose.arn
  }
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}
`, rName)
}

func testAccSubscriptionFilterConfig_kinesisStreamBase(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_cloudwatch_log_group" "test" {
  name              = %[1]q
  retention_in_days = 1
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "logs.${data.aws_region.current.name}.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "kinesis:PutRecord",
      "Resource": "${aws_kinesis_stream.test.arn}"
    },
    {
      "Effect": "Allow",
      "Action": "iam:PassRole",
      "Resource": "${aws_iam_role.test.arn}"
    }
  ]
}
EOF
}

resource "aws_kinesis_stream" "test" {
  name        = %[1]q
  shard_count = 1
}
`, rName)
}

func testAccSubscriptionFilterConfig_lambdaBase(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_cloudwatch_log_group" "test" {
  name              = %[1]q
  retention_in_days = 1
}

resource "aws_iam_role" "test" {
  name = %[1]q

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

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role       = aws_iam_role.test.name
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  runtime       = "nodejs16.x"
  handler       = "exports.handler"
}

resource "aws_lambda_permission" "test" {
  statement_id  = "AllowExecutionFromCloudWatchLogs"
  action        = "lambda:*"
  function_name = aws_lambda_function.test.arn
  principal     = "logs.amazonaws.com"
}
`, rName)
}

func testAccSubscriptionFilterConfig_lambdaMany(rName string, n int) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_cloudwatch_log_group" "test" {
  name              = %[1]q
  retention_in_days = 1
}

resource "aws_iam_role" "test" {
  name = %[1]q

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

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role       = aws_iam_role.test.name
}

resource "aws_lambda_function" "test" {
  count = %[2]d

  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s-${count.index}"
  role          = aws_iam_role.test.arn
  runtime       = "nodejs16.x"
  handler       = "exports.handler"
}

resource "aws_lambda_permission" "test" {
  count = 2

  statement_id  = "AllowExecutionFromCloudWatchLogs"
  action        = "lambda:*"
  function_name = aws_lambda_function.test[count.index].arn
  principal     = "logs.amazonaws.com"
}
`, rName, n)
}

func testAccSubscriptionFilterConfig_destinationARNKinesisDataFirehose(rName string) string {
	return acctest.ConfigCompose(testAccSubscriptionFilterConfig_kinesisDataFirehoseBase(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_subscription_filter" "test" {
  destination_arn = aws_kinesis_firehose_delivery_stream.test.arn
  filter_pattern  = "logtype test"
  log_group_name  = aws_cloudwatch_log_group.test.name
  name            = %[1]q
  role_arn        = aws_iam_role.cloudwatchlogs.arn
}
`, rName))
}

func testAccSubscriptionFilterConfig_destinationARNKinesisStream(rName string) string {
	return acctest.ConfigCompose(testAccSubscriptionFilterConfig_kinesisStreamBase(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_subscription_filter" "test" {
  destination_arn = aws_kinesis_stream.test.arn
  filter_pattern  = "logtype test"
  log_group_name  = aws_cloudwatch_log_group.test.name
  name            = %[1]q
  role_arn        = aws_iam_role.test.arn
}
`, rName))
}

func testAccSubscriptionFilterConfig_destinationARNLambda(rName string) string {
	return acctest.ConfigCompose(testAccSubscriptionFilterConfig_lambdaBase(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_subscription_filter" "test" {
  destination_arn = aws_lambda_function.test.arn
  filter_pattern  = "logtype test"
  log_group_name  = aws_cloudwatch_log_group.test.name
  name            = %[1]q
}
`, rName))
}

func testAccSubscriptionFilterConfig_distribution(rName, distribution string) string {
	return acctest.ConfigCompose(testAccSubscriptionFilterConfig_lambdaBase(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_subscription_filter" "test" {
  destination_arn = aws_lambda_function.test.arn
  distribution    = %[2]q
  filter_pattern  = "logtype test"
  log_group_name  = aws_cloudwatch_log_group.test.name
  name            = %[1]q
}
`, rName, distribution))
}

func testAccSubscriptionFilterConfig_destinationARNLambdaMany(rName string, n int) string {
	return acctest.ConfigCompose(testAccSubscriptionFilterConfig_lambdaMany(rName, n), fmt.Sprintf(`
resource "aws_cloudwatch_log_subscription_filter" "test" {
  count = %[2]d

  destination_arn = aws_lambda_function.test[count.index].arn
  filter_pattern  = "logtype test"
  log_group_name  = aws_cloudwatch_log_group.test.name
  name            = "%[1]s-${count.index}"
}
`, rName, n))
}

func testAccSubscriptionFilterConfig_roleARN1(rName string) string {
	return acctest.ConfigCompose(testAccSubscriptionFilterConfig_kinesisStreamBase(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_subscription_filter" "test" {
  destination_arn = aws_kinesis_stream.test.arn
  filter_pattern  = "logtype test"
  log_group_name  = aws_cloudwatch_log_group.test.name
  name            = %[1]q
  role_arn        = aws_iam_role.test.arn
}
`, rName))
}

func testAccSubscriptionFilterConfig_roleARN2(rName string) string {
	return acctest.ConfigCompose(testAccSubscriptionFilterConfig_kinesisStreamBase(rName), fmt.Sprintf(`
resource "aws_iam_role" "test2" {
  name = "%[1]s-2"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "logs.${data.aws_region.current.name}.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test2" {
  role = aws_iam_role.test2.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "kinesis:PutRecord",
      "Resource": "${aws_kinesis_stream.test.arn}"
    },
    {
      "Effect": "Allow",
      "Action": "iam:PassRole",
      "Resource": "${aws_iam_role.test2.arn}"
    }
  ]
}
EOF
}

resource "aws_cloudwatch_log_subscription_filter" "test" {
  destination_arn = aws_kinesis_stream.test.arn
  filter_pattern  = "logtype test"
  log_group_name  = aws_cloudwatch_log_group.test.name
  name            = %[1]q
  role_arn        = aws_iam_role.test2.arn
}
`, rName))
}
