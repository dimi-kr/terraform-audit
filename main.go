package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform/terraform"
	//tf_state "github.com/hashicorp/terraform/state"
	tf_config "github.com/hashicorp/terraform/config"
	//backend "github.com/hashicorp/terraform/backend"
	backendS3 "github.com/hashicorp/terraform/backend/remote-state/s3"
	"strings"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"reflect"
)

type filterf func(interface{}) bool

func Filter(in interface{}, fn filterf) interface{} {
	val := reflect.ValueOf(in)
	out := make([]interface{}, 0, val.Len())
	for i := 0; i < val.Len(); i++ {
		current := val.Index(i).Interface()
		if fn(current) {
			out = append(out, current)
		}
	}
	return out
}

func stringInSlice(str string, list []string) bool {
for _, v := range list {
if v == str {
return true
}
}
return false
}

func main() {
	tfConfig, err := tf_config.LoadDir("/Users/dimetrio/Desktop/developer-admin-terraform/")
	if err != nil {
		panic("unable to load terraform configuration, " + err.Error())
	}

	backendConfig := tfConfig.Terraform.Backend

	fmt.Println(backendConfig.RawConfig)
	if backendConfig.Type != "s3" {
		fmt.Println("The only supported backend configuration for terraform is 's3'")
	}

	for k,v := range backendConfig.RawConfig.Raw{
		fmt.Println(k, v)
	}

	bc := backendS3.New()
	err = bc.Configure(terraform.NewResourceConfig(backendConfig.RawConfig))
	if err != nil {
		panic("err: %s" + err.Error())
	}
	st,err := bc.States()
	if err != nil {
		panic("err: %s" + err.Error())
	}
	fmt.Println(st)

	state, err := bc.State("prod")
	if err != nil {
		panic("err: %s" + err.Error())
	}

	err = state.RefreshState()
	if err != nil {
		panic("err: %s" + err.Error())
	}

	hasResources := state.State().HasResources()
	if !hasResources {
		panic("No resources found")
	}

	//fmt.Println(state.State().Modules)

	for _, v := range state.State().Modules{
		for rk, rv := range v.Resources {
			fmt.Println("key: ",rk," value: ", rv, rv.Type, rv.Primary.ID, rv.Provider)
		}
	}

	//stateConfig, err := tf_config.NewRawConfig()

	cfg, err := external.LoadDefaultAWSConfig(external.WithSharedConfigProfile("live"))
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}
	cfg.Region = endpoints.EuWest1RegionID

	s3filter := s3.New(cfg)
	s3input := &s3.ListBucketsInput{}
	r := s3filter.ListBucketsRequest(s3input)

	s3resp, err := r.Send()
	if err != nil {
		panic("failed to describe table, " + err.Error())
	}

	filterNames := []string{"shopgate-ami"}

	fmt.Println("Response", Filter(s3resp.Buckets, func(val interface{}) bool {
		return stringInSlice(*val.(s3.Bucket).Name,filterNames)
	}))

	svc := ec2.New(cfg)
	params := &ec2.DescribeInstancesInput{
		Filters: []ec2.Filter{
			{
				Name: aws.String("tag:Name"),
				Values: []string{
					strings.Join([]string{"logs-es"}, ""),
				},
			},
		},
	}

	req := svc.DescribeInstancesRequest(params)

	resp, err := req.Send()
	if err != nil {
		panic("failed to describe table, " + err.Error())
	}
	fmt.Println("Response", resp)
}
