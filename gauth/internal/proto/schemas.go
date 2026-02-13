package proto

// Checkin protobuf schemas translated from MicroG's checkin.proto.
// Field numbers match the .proto definitions exactly.

// CheckinRequestSchema defines the protobuf schema for device check-in.
var CheckinRequestSchema = MessageType{
	"2": {Type: TypeInt},    // androidId (0 on first check-in)
	"3": {Type: TypeString}, // digest
	"4": {Type: TypeMessage, MessageDef: map[string]FieldDef{ // Checkin
		"1": {Type: TypeMessage, MessageDef: map[string]FieldDef{ // Build
			"1":  {Type: TypeString}, // fingerprint
			"2":  {Type: TypeString}, // hardware
			"3":  {Type: TypeString}, // brand
			"4":  {Type: TypeString}, // radio
			"5":  {Type: TypeString}, // bootloader
			"6":  {Type: TypeString}, // clientId
			"7":  {Type: TypeInt},    // time
			"9":  {Type: TypeString}, // device
			"10": {Type: TypeInt},    // sdkVersion
			"11": {Type: TypeString}, // model
			"12": {Type: TypeString}, // manufacturer
			"13": {Type: TypeString}, // product
			"14": {Type: TypeBool},   // otaInstalled
		}},
		"2": {Type: TypeInt}, // lastCheckinMs
		"3": {Type: TypeMessage, Repeated: true, MessageDef: map[string]FieldDef{ // Event
			"1": {Type: TypeString}, // tag
			"2": {Type: TypeString}, // value
			"3": {Type: TypeInt},    // timeMs
		}},
		"6": {Type: TypeString}, // cellOperator
		"7": {Type: TypeString}, // simOperator
		"8": {Type: TypeString}, // roaming
		"9": {Type: TypeInt},    // userNumber
	}},
	"6":  {Type: TypeString},                 // locale
	"7":  {Type: TypeInt},                    // loggingId
	"9":  {Type: TypeString, Repeated: true}, // macAddress
	"10": {Type: TypeString},                 // meid
	"11": {Type: TypeString, Repeated: true}, // accountCookie
	"12": {Type: TypeString},                 // timeZone
	"14": {Type: TypeInt},                    // version (= 3)
	"15": {Type: TypeString, Repeated: true}, // otaCert
	"16": {Type: TypeString},                 // serial
	"17": {Type: TypeString},                 // esn
	"18": {Type: TypeMessage, MessageDef: map[string]FieldDef{ // DeviceConfig
		"1":  {Type: TypeInt},                    // touchScreen
		"2":  {Type: TypeInt},                    // keyboardType
		"3":  {Type: TypeInt},                    // navigation
		"4":  {Type: TypeInt},                    // screenLayout
		"5":  {Type: TypeBool},                   // hasHardKeyboard
		"6":  {Type: TypeBool},                   // hasFiveWayNavigation
		"7":  {Type: TypeInt},                    // densityDpi
		"8":  {Type: TypeInt},                    // glEsVersion
		"9":  {Type: TypeString, Repeated: true}, // sharedLibrary
		"10": {Type: TypeString, Repeated: true}, // availableFeature
		"11": {Type: TypeString, Repeated: true}, // nativePlatform
		"12": {Type: TypeInt},                    // widthPixels
		"13": {Type: TypeInt},                    // heightPixels
		"14": {Type: TypeString, Repeated: true}, // locale
		"15": {Type: TypeString, Repeated: true}, // glExtension
	}},
	"19": {Type: TypeString, Repeated: true}, // macAddressType
	"20": {Type: TypeInt},                    // fragment
	"21": {Type: TypeString},                 // userName
	"22": {Type: TypeInt},                    // userSerialNumber
}
