package models

type RegistrationMail struct {
	Subject    string `json:"subject"`
	PreHeader  string `json:"preHeader"`
	Hello      string `json:"hello"`
	Content    string `json:"content"`
	ButtonText string `json:"buttonText"`
	ButtonUrl  string `json:"buttonUrl"`
	Greetings  string `json:"greetings"`
}
