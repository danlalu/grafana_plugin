import { DataSourceInstanceSettings } from '@grafana/data';
import { DataSourceWithBackend,getBackendSrv } from '@grafana/runtime';
import { MyDataSourceOptions, MyQuery } from './types';
// import {dateMath} from '@grafana/data';
export class DataSource extends DataSourceWithBackend<MyQuery, MyDataSourceOptions> {
  result: any
  url: any
  constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>) {
    super(instanceSettings);
    this.result = null;
    // this.username = instanceSettings.jsonData.username;
  }
  nodeQuery(query: any, options?: any): any {
    return this.getChildPaths(query);
  }
  async getChildPaths(params: any) {
    // if (this.url.substr(this.url.length - 1, 1) === '/') {
    //   this.url = this.url.substr(0, this.url.length - 1);
    // }
    // return this.postResource('query', params)
    //   .then((response) => {
    //     if (response instanceof Array) {
    //       return response;
    //     } else {
    //       throw 'the result is not array';
    //     }
    //   })
      // .then((data) => data.map(toMetricFindValue));

      return getBackendSrv().post('/api/ds/query', params)
  }
  createAlert(query: any, options?: any): any {
    return this.createAlertFun(query,options);
  }
  async createAlertFun(params: any,options: any) {
    return getBackendSrv().post(`/api/datasources/${options.id}/resources/createAlert`, params)
  }
  removeAlert(query: any, options?: any): any {
    return this.removeAlertFun(query,options);
  }
  async removeAlertFun(params: any,options: any) {
    return getBackendSrv().post(`/api/datasources/${options.id}/resources/removeAlert`, params)
  }
  realtimeTaskSave(query: any, options?: any): any {
    return this.realtimeTaskSaveFun(query,options);
  }
  async realtimeTaskSaveFun(params: any,options: any) {
    return getBackendSrv().post(`/api/datasources/${options.id}/resources/realtimeTaskSave`, params)
  }
  realtimeTaskRemove(query: any, options?: any): any {
    return this.realtimeTaskRemoveFun(query,options);
  }
  async realtimeTaskRemoveFun(params: any,options: any) {
    return getBackendSrv().post(`/api/datasources/${options.id}/resources/realtimeTaskRemove`, params)
  }
  algorithmList(query: any, options?: any): any {
    return this.algorithmListFun(query,options);
  }
  async algorithmListFun(params: any,options: any) {
    return getBackendSrv().post(`/api/datasources/${options.id}/resources/algorithmList`, params)
  }
  getMetrics(query: any, options?: any): any {
    return this.getMetricsFun(query,options);
  }
  async getMetricsFun(params: any, options: any) {
    return getBackendSrv().post(`/api/datasources/${options.id}/resources/metrics`, params)
  }
  getLabelNames (query: any, options?: any): any {
    return this.getLabelNamesFun(query,options);
  }
  async getLabelNamesFun(params: any, options: any) {
    return getBackendSrv().post(`/api/datasources/${options.id}/resources/labelNames`, params)
  }
  getSeries(query: any, options?: any): any {
    return this.getSeriesFun(query,options);
  }
  async getSeriesFun(params: any, options: any) {
    return getBackendSrv().post(`/api/datasources/${options.id}/resources/series`, params)
  }
  // getPrometheusTime(date: any, roundUp: boolean) {
  //   if (typeof date === 'string') {
  //     date = dateMath.parse(date, roundUp)!;
  //   }

  //   return Math.ceil(date.valueOf() / 1000);
  // }

  // getTimeRangeParams(): { start: string; end: string } {
  //   const range = this.timeSrv.timeRange();
  //   return {
  //     start: this.getPrometheusTime(range.from, false).toString(),
  //     end: this.getPrometheusTime(range.to, true).toString(),
  //   };
  // }



}
