package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/joho/godotenv"
)

func NewDutaStack(scope constructs.Construct, id string, props *awscdk.StackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, props)

	awsdynamodb.NewTable(stack, jsii.String("Workspaces"), &awsdynamodb.TableProps{
		TableName:     jsii.String("duta-workspaces"),
		PartitionKey:  &awsdynamodb.Attribute{Name: jsii.String("PK"), Type: awsdynamodb.AttributeType_STRING},
		SortKey:       &awsdynamodb.Attribute{Name: jsii.String("SK"), Type: awsdynamodb.AttributeType_STRING},
		BillingMode:   awsdynamodb.BillingMode_PAY_PER_REQUEST,
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	return stack
}

func main() {
	defer jsii.Close()

	_ = godotenv.Load()

	app := awscdk.NewApp(nil)
	NewDutaStack(app, "DutaStack", &awscdk.StackProps{
		Env: &awscdk.Environment{Region: jsii.String(os.Getenv("AWS_REGION"))},
	})
	app.Synth(nil)
}
