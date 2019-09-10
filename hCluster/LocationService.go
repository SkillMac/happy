package hCluster

import "custom/happy/hLog"

type LocationService struct {
	location *LocationComponent
}

func (this *LocationService) init(mlocation *LocationComponent) {
	this.location = mlocation
}

func (this *LocationService) NodeInquiry(args []string, reply *[]*InquiryReply) error {
	hLog.Debug("Inquiry :", args)
	res, err := this.location.NodeInquiry(args, false)
	*reply = res
	return err
}

func (this *LocationService) NodeInquiryDetail(args []string, reply *[]*InquiryReply) error {
	res, err := this.location.NodeInquiry(args, true)
	*reply = res
	return err
}

func (this *LocationService) NodeLogInquiry(args int64, reply *[]*NodeLog) error {
	res, err := this.location.NodeLogInquiry(args)
	*reply = res
	return err
}
