package nsfContext

import (
	"encoding/json"
	"fmt"
	"nsf/internal/db"
	"nsf/logger"
)

type serviceInfo struct {
	NfName     string `json:"nf_name"`
	SbiAddress string `json:"sbi_address"`
	SbiPort    int    `json:"sbi_port"`
}

// TODO: You may need to add some fields
type serviceContextIE struct {
	db.DbIf

	*logger.NsfLogger
}

// TODO: You may need to add some fields
type serviceContext struct {
	db.DbIf

	*logger.NsfLogger
}

// TODO: Finish the initialization of the serviceContext instance
func newServiceContext(serviceContextIE *serviceContextIE) *serviceContext {

	return &serviceContext{
		DbIf: serviceContextIE.DbIf,

		NsfLogger: serviceContextIE.NsfLogger,
	}
}

// TODO: Finish the release of the serviceContext instance
func (s *serviceContext) release() {
	s.NsfLogger.SvcLog.Infoln("Release serviceContext...")

	s.NsfLogger.SvcLog.Infoln("serviceContext released")
}

// TODO: Finish the implementation of the registerServices method with an error return type
func (s *serviceContext) registerServices(nfName, sbiAddr string, sbiPort int, services []string) error {
	s.NsfLogger.SvcLog.Tracef("Registering %s services", nfName)

	info := serviceInfo{
		NfName:     nfName,
		SbiAddress: sbiAddr,
		SbiPort:    sbiPort,
	}

	jsonData, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to encode service information: %w", err)
	}

	value := string(jsonData)

	data := make(map[string]string, len(services))
	for _, service := range services {
		data[service] = value
	}

	if err := s.SaveAll(data); err != nil {
		return fmt.Errorf("failed to register services for NF %s: %w", nfName, err)
	}

	return nil
}

// TODO: Finish the implementation of the deregisterServices method with an error return type
func (s *serviceContext) deregisterServices(nfName string, services []string) error {
	s.NsfLogger.SvcLog.Tracef("De-registering %s services", nfName)
	for _, service := range services {
		// 取得資料庫記錄的提供者名字
		registeredNfName, _, _, err := s.discoverService(service)
		if err != nil {
			return fmt.Errorf("failed to deregister service %s: %w", service, err)
		}

		// 2. 比對請求中的 NF 名字與資料庫中的 NF 名字。
		if registeredNfName != nfName {
			return fmt.Errorf(
				"service %s does not belong to NF %s",
				service,
				nfName,
			)
		}

		// 3. 相符才刪除，其他服務的紀錄不受影響。
		if err := s.Remove(service); err != nil {
			return fmt.Errorf("failed to remove service %s: %w", service, err)
		}
	}

	return nil
}

// TODO: Finish the implementation of the discoverService method with these return values:
// (string, string, int, error) representing the NF name, SBI address, SBI port, and an error if any
func (s *serviceContext) discoverService(service string) (string, string, int, error) {
	s.NsfLogger.SvcLog.Tracef("Discovering service %s", service)
	//確認服務是否存在。
	exists, err := s.Exist(service)
	if err != nil {
		return "", "", 0, fmt.Errorf(
			"failed to check service %s: %w",
			service,
			err,
		)
	}
	if !exists {
		return "", "", 0, fmt.Errorf("service %s not found", service)
	}

	// 2. 讀出資料庫裡的 JSON 字串。
	value, err := s.Load(service)
	if err != nil {
		return "", "", 0, fmt.Errorf(
			"failed to load service %s: %w",
			service,
			err,
		)
	}

	// 3. 將 JSON 解析回 Go struct。
	var info serviceInfo
	if err := json.Unmarshal([]byte(value), &info); err != nil {
		return "", "", 0, fmt.Errorf(
			"failed to decode service %s: %w",
			service,
			err,
		)
	}

	// 4. 成功時回傳三個欄位，以及 nil。
	return info.NfName, info.SbiAddress, info.SbiPort, nil

}
