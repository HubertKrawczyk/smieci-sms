package service

type SMSService interface {
    SendSMS(phone, message string) error
}

type smsService struct {
    apiKey string
}

func NewSMSService(apiKey string) SMSService {
    return &smsService{apiKey: apiKey}
}

func (s *smsService) SendSMS(phone, message string) error {
    // TODO: integrate with SMS provider (Twilio, Nexmo, etc.)
    return nil
}
