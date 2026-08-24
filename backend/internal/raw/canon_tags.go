package raw

var canonShotInfo = map[uint16]string{
	0x0001: "CanonMacroMode",
	0x0002: "CanonSelfTimer",
	0x0003: "CanonQuality",
	0x0004: "CanonCanonFlashMode",
	0x0005: "CanonContinuousDrive",
	0x0007: "CanonFocusMode",
	0x0009: "CanonRecordMode",
	0x000A: "CanonCanonImageSize",
	0x000B: "CanonEasyMode",
	0x000C: "CanonDigitalZoom",
	0x000D: "CanonContrast",
	0x000E: "CanonSaturation",
	0x000F: "CanonSharpness",
	0x0010: "CanonCameraISO",
	0x0011: "CanonMeteringMode",
	0x0012: "CanonFocusType",
	0x0013: "CanonAFPoint",
	0x0014: "CanonExposureProgram",
	0x0016: "CanonLensType",
	0x0017: "CanonMaxFocalLength",
	0x0018: "CanonMinFocalLength",
	0x0019: "CanonFocalUnits",
	0x001A: "CanonMaxAperture",
	0x001B: "CanonMinAperture",
	0x001C: "CanonFlashActivity",
	0x001D: "CanonFlashBits",
	0x001E: "CanonFocusContinuous",
	0x001F: "CanonAESetting",
	0x0020: "CanonImageStabilization",
	0x0021: "CanonDisplayAperture",
	0x0022: "CanonZoomSourceWidth",
	0x0023: "CanonZoomTargetWidth",
	0x0025: "CanonSpotMeteringMode",
	0x0026: "CanonPhotoEffect",
	0x0027: "CanonManualFlashOutput",
	0x0028: "CanonColorTone",
	0x002D: "CanonSRAWQuality",
}

var canonCameraSettings = map[uint16]string{
	0x0001: "CanonExposureMode",
	0x0095: "CanonLensModel",
	0x0096: "CanonSerialNumber",
	0x0098: "CanonLensInfo",
}

func canonTagName(id uint16) string {
	if n, ok := canonShotInfo[id]; ok {
		return n
	}
	if n, ok := canonCameraSettings[id]; ok {
		return n
	}
	return ""
}
