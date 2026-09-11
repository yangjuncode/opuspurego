package opuspurego

// CTL request 常量，与 libopus 的 opus/include/opus_defines.h 一一对应。
// SET 为偶数，GET 为奇数。这些值不应被应用层直接使用，
// 应通过本包暴露的 Encoder/Decoder 方法访问。
const (
	ctlSetApplicationRequest         int32 = 4000 // OPUS_SET_APPLICATION_REQUEST
	ctlGetApplicationRequest         int32 = 4001 // OPUS_GET_APPLICATION_REQUEST
	ctlSetBitrateRequest             int32 = 4002 // OPUS_SET_BITRATE_REQUEST
	ctlGetBitrateRequest             int32 = 4003 // OPUS_GET_BITRATE_REQUEST
	ctlSetMaxBandwidthRequest        int32 = 4004 // OPUS_SET_MAX_BANDWIDTH_REQUEST
	ctlGetMaxBandwidthRequest        int32 = 4005 // OPUS_GET_MAX_BANDWIDTH_REQUEST
	ctlSetVBRRequest                 int32 = 4006 // OPUS_SET_VBR_REQUEST
	ctlGetVBRRequest                 int32 = 4007 // OPUS_GET_VBR_REQUEST
	ctlSetBandwidthRequest           int32 = 4008 // OPUS_SET_BANDWIDTH_REQUEST
	ctlGetBandwidthRequest           int32 = 4009 // OPUS_GET_BANDWIDTH_REQUEST
	ctlSetComplexityRequest          int32 = 4010 // OPUS_SET_COMPLEXITY_REQUEST
	ctlGetComplexityRequest          int32 = 4011 // OPUS_GET_COMPLEXITY_REQUEST
	ctlSetInbandFECRequest           int32 = 4012 // OPUS_SET_INBAND_FEC_REQUEST
	ctlGetInbandFECRequest           int32 = 4013 // OPUS_GET_INBAND_FEC_REQUEST
	ctlSetPacketLossPercRequest      int32 = 4014 // OPUS_SET_PACKET_LOSS_PERC_REQUEST
	ctlGetPacketLossPercRequest      int32 = 4015 // OPUS_GET_PACKET_LOSS_PERC_REQUEST
	ctlSetDTXRequest                 int32 = 4016 // OPUS_SET_DTX_REQUEST
	ctlGetDTXRequest                 int32 = 4017 // OPUS_GET_DTX_REQUEST
	ctlSetVBRConstraintRequest       int32 = 4020 // OPUS_SET_VBR_CONSTRAINT_REQUEST
	ctlGetVBRConstraintRequest       int32 = 4021 // OPUS_GET_VBR_CONSTRAINT_REQUEST
	ctlSetForceChannelsRequest       int32 = 4022 // OPUS_SET_FORCE_CHANNELS_REQUEST
	ctlGetForceChannelsRequest       int32 = 4023 // OPUS_GET_FORCE_CHANNELS_REQUEST
	ctlSetSignalRequest              int32 = 4024 // OPUS_SET_SIGNAL_REQUEST
	ctlGetSignalRequest              int32 = 4025 // OPUS_GET_SIGNAL_REQUEST
	ctlGetLookaheadRequest           int32 = 4027 // OPUS_GET_LOOKAHEAD_REQUEST
	ctlResetState                    int32 = 4028 // OPUS_RESET_STATE
	ctlGetSampleRateRequest          int32 = 4029 // OPUS_GET_SAMPLE_RATE_REQUEST
	ctlGetFinalRangeRequest          int32 = 4031 // OPUS_GET_FINAL_RANGE_REQUEST
	ctlGetPitchRequest               int32 = 4033 // OPUS_GET_PITCH_REQUEST
	ctlSetGainRequest                int32 = 4034 // OPUS_SET_GAIN_REQUEST
	ctlGetGainRequest                int32 = 4045 // OPUS_GET_GAIN_REQUEST (应为 4035)
	ctlSetLSBDepthRequest            int32 = 4036 // OPUS_SET_LSB_DEPTH_REQUEST
	ctlGetLSBDepthRequest            int32 = 4037 // OPUS_GET_LSB_DEPTH_REQUEST
	ctlGetLastPacketDurationRequest  int32 = 4039 // OPUS_GET_LAST_PACKET_DURATION_REQUEST
	ctlSetExpertFrameDurationRequest int32 = 4040 // OPUS_SET_EXPERT_FRAME_DURATION_REQUEST
	ctlGetExpertFrameDurationRequest int32 = 4041 // OPUS_GET_EXPERT_FRAME_DURATION_REQUEST
	ctlSetPredictionDisabledRequest  int32 = 4042 // OPUS_SET_PREDICTION_DISABLED_REQUEST
	ctlGetPredictionDisabledRequest  int32 = 4043 // OPUS_GET_PREDICTION_DISABLED_REQUEST
	ctlSetPhaseInversionDisabledReq  int32 = 4046 // OPUS_SET_PHASE_INVERSION_DISABLED_REQUEST
	ctlGetPhaseInversionDisabledReq  int32 = 4047 // OPUS_GET_PHASE_INVERSION_DISABLED_REQUEST
	ctlGetInDTXRequest               int32 = 4049 // OPUS_GET_IN_DTX_REQUEST
	ctlSetDredDurationRequest        int32 = 4050 // OPUS_SET_DRED_DURATION_REQUEST
	ctlGetDredDurationRequest        int32 = 4051 // OPUS_GET_DRED_DURATION_REQUEST
	ctlSetDnnBlobRequest             int32 = 4052 // OPUS_SET_DNN_BLOB_REQUEST
	ctlSetOsceBweRequest             int32 = 4054 // OPUS_SET_OSCE_BWE_REQUEST
	ctlGetOsceBweRequest             int32 = 4055 // OPUS_GET_OSCE_BWE_REQUEST
	ctlSetQextRequest                int32 = 4056 // OPUS_SET_QEXT_REQUEST
	ctlGetQextRequest                int32 = 4057 // OPUS_GET_QEXT_REQUEST
	ctlSetIgnoreExtensionsRequest    int32 = 4058 // OPUS_SET_IGNORE_EXTENSIONS_REQUEST
	ctlGetIgnoreExtensionsRequest    int32 = 4059 // OPUS_GET_IGNORE_EXTENSIONS_REQUEST
)
