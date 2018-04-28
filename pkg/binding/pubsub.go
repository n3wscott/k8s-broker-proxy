package binding

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

const (
	gcpProjectEnvName   = "GOOGLE_CLOUD_PROJECT"
	pubsubTopicEnvName  = "PUBSUB_TOPIC"
	pubsubSubEnvName    = "PUBSUB_SUBSCRIPTION"
	gcpApplicationCreds = "GOOGLE_APPLICATION_CREDENTIALS"
)

type PubSubBindingObject struct {
	PrivateKeyData string `json:"privateKeyData"`
	ProjectId      string `json:"projectId"`
	SubscriptionId string `json:"subscriptionId"`
	TopicId        string `json:"topicId"`
}

func PubSubBinding(file string) (projectId, topicId, subscriptionId string, err error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	var binding map[string]*json.RawMessage

	err = json.Unmarshal(data, &binding)
	if err != nil {
		return
	}

	var pubsubBinding PubSubBindingObject
	json.Unmarshal(*binding["data"], &pubsubBinding)
	if err != nil {
		return
	}

	if err != nil {
		return
	}

	projectIdBytes, err := base64.StdEncoding.DecodeString(pubsubBinding.ProjectId)
	if err != nil {
		return
	}
	projectId = string(projectIdBytes)

	topicIdBytes, err := base64.StdEncoding.DecodeString(pubsubBinding.TopicId)
	if err != nil {
		return
	}
	topicId = string(topicIdBytes)

	subscriptionIdBytes, err := base64.StdEncoding.DecodeString(pubsubBinding.SubscriptionId)
	if err != nil {
		return
	}
	subscriptionId = string(subscriptionIdBytes)

	credsFile, err := ioutil.TempFile(os.TempDir(), "ledhouse")

	creds, err := base64.StdEncoding.DecodeString(pubsubBinding.PrivateKeyData)
	if err != nil {
		fmt.Println("decode error:", err)
		return
	}
	credsFile.Write(creds)

	os.Setenv(gcpProjectEnvName, projectId)
	os.Setenv(pubsubTopicEnvName, topicId)
	os.Setenv(pubsubSubEnvName, subscriptionId)
	os.Setenv(gcpApplicationCreds, credsFile.Name())
	return
}
