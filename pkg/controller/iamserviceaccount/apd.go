package iamserviceaccount

type condition struct {
	StringEquals map[string]string
}

type principal struct {
	Federated string
}

type statement struct {
	Effect    string
	Principal principal
	Action    string
	Condition condition
}

type assumePolicyDocument struct {
	Version   string
	Statement []statement
}
