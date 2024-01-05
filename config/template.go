package config

func GenerateTemplate(t string) string {
	switch t {
	case "PROTO":
		return `{
			"distributed-cache": {
			  "mode": "SYNC",
			  "encoding": {
				"media-type": "application/x-protostream"
			  },
			  "statistics": true
			}
		  }`
	case "JBOSS":
		return `{
				"distributed-cache": {
				  "mode": "SYNC",
				  "encoding": {
					"media-type": "application/x-jboss-marshalling"
				  },
				  "statistics": true
				}
			  }`
	default:
		return ""
	}
}
