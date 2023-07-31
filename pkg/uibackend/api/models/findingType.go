package models

func GetFindingTypes() []FindingType {
	return []FindingType{
		EXPLOIT,
		MALWARE,
		MISCONFIGURATION,
		PACKAGE,
		ROOTKIT,
		SECRET,
		VULNERABILITY,
	}
}
