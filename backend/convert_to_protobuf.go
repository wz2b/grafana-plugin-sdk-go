package backend

import (
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/dataframe"
	"github.com/grafana/grafana-plugin-sdk-go/genproto/pluginv2"
)

type convertToProtobuf struct {
}

func toProto() convertToProtobuf {
	return convertToProtobuf{}
}

func (t convertToProtobuf) PluginConfig(config PluginConfig) *pluginv2.PluginConfig {
	protoConfig := &pluginv2.PluginConfig{
		OrgId:      config.OrgID,
		PluginId:   config.PluginID,
		PluginType: config.PluginType,
	}

	if config.AppSettings != nil {
		protoConfig.InstanceSettings = t.AppInstanceSettings(config.AppSettings)
	}

	if config.DataSourceSettings != nil {
		protoConfig.InstanceSettings = t.DataSourceInstanceSettings(config.DataSourceSettings)
	}

	return protoConfig
}

func (t convertToProtobuf) AppInstanceSettings(settings *AppInstanceSettings) *pluginv2.PluginConfig_App {
	if settings == nil {
		return nil
	}

	return &pluginv2.PluginConfig_App{
		App: &pluginv2.PluginConfig_AppInstanceSettings{
			JsonData:                settings.JSONData,
			DecryptedSecureJsonData: settings.DecryptedSecureJSONData,
			UpdatedMS:               settings.Updated.UnixNano() / int64(time.Millisecond),
		},
	}
}

func (t convertToProtobuf) DataSourceInstanceSettings(settings *DataSourceInstanceSettings) *pluginv2.PluginConfig_DataSource {
	if settings == nil {
		return nil
	}

	return &pluginv2.PluginConfig_DataSource{
		DataSource: &pluginv2.PluginConfig_DataSourceInstanceSettings{
			Id:                      settings.ID,
			Name:                    settings.Name,
			Url:                     settings.URL,
			User:                    settings.User,
			Database:                settings.Database,
			BasicAuthEnabled:        settings.BasicAuthEnabled,
			BasicAuthUser:           settings.BasicAuthUser,
			JsonData:                settings.JSONData,
			DecryptedSecureJsonData: settings.DecryptedSecureJSONData,
			UpdatedMS:               settings.Updated.UnixNano() / int64(time.Millisecond),
		},
	}
}

func (t convertToProtobuf) TimeRange(tr TimeRange) *pluginv2.TimeRange {
	return &pluginv2.TimeRange{
		FromEpochMS: tr.From.UnixNano() / int64(time.Millisecond),
		ToEpochMS:   tr.To.UnixNano() / int64(time.Millisecond),
	}
}

func (t convertToProtobuf) HealthStatus(status HealthStatus) pluginv2.CheckHealth_Response_HealthStatus {
	switch status {
	case HealthStatusUnknown:
		return pluginv2.CheckHealth_Response_UNKNOWN
	case HealthStatusOk:
		return pluginv2.CheckHealth_Response_OK
	case HealthStatusError:
		return pluginv2.CheckHealth_Response_ERROR
	}
	panic("unsupported protobuf health status type in sdk")
}

func (t convertToProtobuf) CheckHealthResponse(res *CheckHealthResult) *pluginv2.CheckHealth_Response {
	return &pluginv2.CheckHealth_Response{
		Status: t.HealthStatus(res.Status),
		Info:   res.Info,
	}
}

func (t convertToProtobuf) DataQuery(q DataQuery) *pluginv2.DataQuery {
	return &pluginv2.DataQuery{
		RefId:         q.RefID,
		MaxDataPoints: q.MaxDataPoints,
		IntervalMS:    q.Interval.Milliseconds(),
		TimeRange:     t.TimeRange(q.TimeRange),
		Json:          q.JSON,
	}
}

func (t convertToProtobuf) DataQueryRequest(req *DataQueryRequest) *pluginv2.DataQueryRequest {
	queries := make([]*pluginv2.DataQuery, len(req.Queries))
	for i, q := range req.Queries {
		queries[i] = t.DataQuery(q)
	}
	return &pluginv2.DataQueryRequest{
		Config:  t.PluginConfig(req.PluginConfig),
		Headers: req.Headers,
		Queries: queries,
	}
}

func (t convertToProtobuf) DataQueryResponse(res *DataQueryResponse) (*pluginv2.DataQueryResponse, error) {
	encodedFrames := make([][]byte, len(res.Frames))
	var err error
	for i, frame := range res.Frames {
		encodedFrames[i], err = dataframe.MarshalArrow(frame)
		if err != nil {
			return nil, err
		}
	}

	return &pluginv2.DataQueryResponse{
		Frames:   encodedFrames,
		Metadata: res.Metadata,
	}, nil
}

func (t convertToProtobuf) CallResourceResponse(resp *CallResourceResponse) *pluginv2.CallResource_Response {
	headers := map[string]*pluginv2.CallResource_StringList{}

	for key, values := range resp.Headers {
		headers[key] = &pluginv2.CallResource_StringList{Values: values}
	}

	return &pluginv2.CallResource_Response{
		Headers: headers,
		Code:    int32(resp.Status),
		Body:    resp.Body,
	}
}