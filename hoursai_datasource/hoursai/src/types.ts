import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface MyQuery extends DataQuery {
  expr: string;
  genericId: string;
  queryType: string;
  algorithmList: boolean;
  A_Realtime_Save: string;
  alertTemplateId: number;
  alertEnable: boolean;
  alertTitle: any,
  taskId: any,
  series: any,
}

export const defaultQuery: Partial<MyQuery> = {
  expr: "",
  genericId:"-1",
  queryType: "syncPreview",
  algorithmList:false,
  A_Realtime_Save:"",
  alertTemplateId:1,
  alertEnable: false,
  alertTitle: "1",
  taskId: "",
  series: "",
};

/**
 * These are options configured for each DataSource instance.
 */
export interface MyDataSourceOptions extends DataSourceJsonData {
 // path?: string;
  managerUrl: string;
  token: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
// export interface MySecureJsonData {
//   apiKey?: string;
// }
