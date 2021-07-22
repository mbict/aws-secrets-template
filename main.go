package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"gopkg.in/yaml.v3"
	"os"
	"text/template"
)

type Config struct {
	Templates []struct {
		Secret   string `yaml:"secret"`
		Template string `yaml:"template"`
	} `yaml:"templates"`
}

var (
	yamlFile string
	destFile string
	region   string
)

func init() {
	flag.StringVar(&yamlFile, "c", "", "template configutation")
	flag.StringVar(&destFile, "o", "", "output file to store the config")
	flag.StringVar(&region, "r", "eu-west-1", "aws region")
}

func main() {
	flag.Parse()

	if yamlFile == "" {
		panic("no configuration file defined")
		flag.Usage()
	}

	//load config
	var config Config
	cf, err := os.Open(yamlFile)
	must(err)
	defer cf.Close()
	decoder := yaml.NewDecoder(cf)
	must(decoder.Decode(&config))

	//aws secrets manager
	awsSess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	svc := secretsmanager.New(awsSess)

	//open destination file
	dst := os.Stdout
	if destFile != "" {
		dst, err = os.Create(destFile)
		must(err)
		defer dst.Close()
	}

	for _, secretConfig := range config.Templates {
		//parse template
		t := template.New("config").Funcs(sprig.GenericFuncMap())
		template.Must(t.Parse(secretConfig.Template))

		//get secrets from aws
		result, err := svc.GetSecretValue(&secretsmanager.GetSecretValueInput{
			SecretId: aws.String(secretConfig.Secret),
		})
		must(err)

		//unmarshal data
		secretsData := make(map[string]interface{})
		must(json.Unmarshal([]byte(*result.SecretString), &secretsData))

		//output template to standard output
		must(t.Execute(dst, secretsData))
	}

	if dst != os.Stdout {
		fmt.Println("secrets template file written")
	}
}

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}
