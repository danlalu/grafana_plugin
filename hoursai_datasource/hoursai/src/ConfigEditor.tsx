import React, { ChangeEvent, PureComponent } from 'react';
import { LegacyForms } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { MyDataSourceOptions } from './types';

const { FormField } = LegacyForms;

interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions> {}

interface State {}

export class ConfigEditor extends PureComponent<Props, State> {
  // onPathChange = (event: ChangeEvent<HTMLInputElement>) => {
  //   const { onOptionsChange, options } = this.props;
  //   const jsonData = {
  //     ...options.jsonData,
  //     path: event.target.value,
  //   };
  //   onOptionsChange({ ...options, jsonData });
  // };
  onManagerUrlChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      managerUrl: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };
  onTokenChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      token: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };
  onURLChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({ ...options, url:event.target.value });
  };

  // Secure field (only sent to the backend)
  onAPIKeyChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonData: {
        apiKey: event.target.value,
      },
    });
  };

  onResetAPIKey = () => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      // secureJsonFields: {
      //   ...options.secureJsonFields,
      //   apiKey: false,
      // },
      secureJsonData: {
        ...options.secureJsonData,
        apiKey: '',
      },
    });
  };

  render() {
    const { options } = this.props;
    //const { jsonData, secureJsonFields,url } = options;
    const { jsonData,url } = options;
    // const secureJsonData = (options.secureJsonData || {}) as MySecureJsonData;

    return (
      <div className="gf-form-group">
         <div className="gf-form">
          <FormField
            label="URL"
            labelWidth={6}
            inputWidth={20}
            onChange={this.onURLChange}
            value={url || ''}
            placeholder="请输入URL"
          />
        </div>
        <div className="gf-form">
          <FormField
            label="HoursAIUrl"
            labelWidth={6}
            inputWidth={20}
            onChange={this.onManagerUrlChange}
            value={jsonData.managerUrl || ''}
            placeholder="请输入HoursAIUrl"
          />
        </div>
        {/* <div className="gf-form">
          <FormField
            label="Path"
            labelWidth={6}
            inputWidth={20}
            onChange={this.onPathChange}
            value={jsonData.path || ''}
            placeholder="json field returned to frontend"
          />
        </div> */}

        {/* <div className="gf-form-inline">
          <div className="gf-form">
            <SecretFormField
              isConfigured={(secureJsonFields && secureJsonFields.apiKey) as boolean}
              value={secureJsonData.apiKey || ''}
              label="API Key"
              placeholder="secure json field (backend only)"
              labelWidth={6}
              inputWidth={20}
              onReset={this.onResetAPIKey}
              onChange={this.onAPIKeyChange}
            />
          </div>
        </div> */}
        <div className="gf-form">
          <FormField
            label="Token"
            labelWidth={6}
            inputWidth={20}
            onChange={this.onTokenChange}
            value={jsonData.token || ''}
            placeholder="请输入Token"
          />
        </div>
      </div>
    );
  }
}
