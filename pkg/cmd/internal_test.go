package cmd

import (
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
)

type MockECSAPI struct {
	ecsiface.ECSAPI    // embedding of the interface is needed to skip implementation of all methods
	ListClustersMethod func(input *ecs.ListClustersInput) (*ecs.ListClustersOutput, error)
}

func (m *MockECSAPI) ListClusters(input *ecs.ListClustersInput) (*ecs.ListClustersOutput, error) { // This allows the test to use the same method
	if m.ListClustersMethod != nil {
		return m.ListClustersMethod(input) // We intercept and return a made up reply
	}
	return nil, nil // return any value you think is good for you
}

func TestGetCluster(t *testing.T) {
	mockEcs := &MockECSAPI{
		ListClustersMethod: func(input *ecs.ListClustersInput) (*ecs.ListClustersOutput, error) {
			return &ecs.ListClustersOutput{
				ClusterArns: []*string{
					//aws.String("arn:aws:ecs:eu-west-1:1111111111:cluster/execCommand"),
					//aws.String("arn:aws:ecs:eu-west-1:1111111111:cluster/bluegreen"),
				},
			}, nil
		},
	}
	clusterName, err := getCluster(mockEcs)
	log.Println(clusterName, err)
}
