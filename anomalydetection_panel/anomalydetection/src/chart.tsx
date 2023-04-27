// todo 在dashborad加载插件的时候 应该优先处理数据的配色，应该会解决tooltip的配色问题
import React, {useEffect, useMemo, useState} from 'react';
// import { getDashboardSrv } from 'app/features/dashboard/services/DashboardSrv';
// import {RefreshEvent, config, DataSourcePicker} from '@grafana/runtime'
// import { updateLocation } from 'app/core/actions';
import { TooltipDisplayMode } from '@grafana/schema';
// import { config } from 'grafana/app/core/config';
import _ from 'lodash';
import * as UI from '@grafana/ui';
import {dateTimeParse} from '@grafana/data' 
// import TimeSeries from 'grafana/app/core/time_series';
import { usePanelContext, TimeSeries, TooltipPlugin, ZoomPlugin, KeyboardPlugin } from '@grafana/ui';
import { getTimezones, prepareGraphableFields, regenerateLinksSupplier, initData } from './utils';
// import moment from 'moment'
const COME_FROM_DASHBOARD = 'dashboard'
const pickAndBuild = (frames: any, options: any) => {
  const {BoundaryWithLenged} = options;
  let showOpts = BoundaryWithLenged ? {} : {
    ...options
  }
  if(!frames?.length) {
    return [];
  }
  frames.forEach((item: any) => {
    if(!item.uuid) {
      item.comeFrom = COME_FROM_DASHBOARD
    }
  })
  const dashboardFlag = frames.every((item: any) => {
    return !item.uuid || item.comeFrom === COME_FROM_DASHBOARD
  })
  // console.log('=====>dashboardFlag', dashboardFlag);
  const baseData = frames.filter((item: any) => !['lower', 'upper', 'baseline', 'anomaly', 'significance'].includes(item.name))
  if(dashboardFlag) { // 从面板看的
    dashboardFlag && initData(frames)
    const showItem = baseData.filter((item: any) => item?.fields[1]?.config?.custom?.hideFrom?.viz === false)
    if(showItem.length === 1) {
      showOpts = {
       [`${showItem[0].labelKey}Show`]: true
      }
    }
  }
  // let _index = -1
  const showed = frames.filter((item: any) => {
    return !item.fields[1].config.custom?.hideFrom?.viz && /^\d+$/.test(item.uuid)
  })
  if( BoundaryWithLenged) {
    showOpts = showed.reduce((pre: any, next: any) => {
      return {
        ...pre,
        [`${next.labelKey}Show`]: true
      }
    }, {})
  }
  if(showed.length === baseData.length) {
    showOpts = {}
  }
  frames.forEach((item: any) => {
    if(/\-Lower$|\-Upper$|\-anomaly$/g.test(item.uuid)) {
      const originItem = frames.filter((frame: any) => {
        return frame.uuid === item.originUuid
      })

      item.fields[1].config.displayNameFromDS =  originItem[0].name + '-' + item.originName;

      const targetOne = frames.filter((frame: any) => {
        if(item.uuid.slice(-5) === 'Upper') {
          return frame.uuid === `${item.originUuid}-Lower`
        } else if(item.uuid.slice(-5) === 'Lower'){
          return frame.uuid === `${item.originUuid}-Upper`
        }
        return false
      })
      // console.log('=====>targetOne', targetOne);
      if(targetOne.length) {
        // console.log('=====>', item.labelKey, showOpts[`${item.labelKey}Show`]);
        item.fields[1].config.custom = {
          fillBelowTo: targetOne[0].fields[1].name,
          lineWidth: 0.1,
          showPoints: 'auto' || 'always',
          pointSize: 5,
          hideFrom: {
            "legend": true,
            "viz": !showOpts[`${item.labelKey}Show`]
          }
        }
      }
      if(/-anomaly$/g.test(item.uuid)) {
        // const time = item.fields[0].values.toArray();
        // const value = item.fields[1].values.toArray();
        // let targetTime: any = [];
        // let targetValue: any = [];
        // value.forEach((_value: any, index: any) => {
        //   // if(_value !== 0 || index === 100) {
        //     targetTime.push(time[index])
        //     targetValue.push(_value === 0 ? null  : _value)
        //   // } 
        // })
        // item.fields[0].values = new ArrayVector(targetTime)
        // item.fields[1].values = new ArrayVector(targetValue)
        item.fields[1].config.color = {
          mode: 'fixed',
            fixedColor: 'red'
        }
        item.fields[1].config.custom = {
          // fillBelowTo: targetOne[0].fields[1].name,
          lineWidth: 0,
          showPoints: 'always',
          pointSize: 4,
          hideFrom: {
            "legend": true,
            "viz": !showOpts[`${item.labelKey}Show`],
            "tooltip": true
          }
        }
      }
    } else if(item.originName === 'significance' || item.originName === 'baseline' ) {
      item.fields[1].config.custom = {
        // fillBelowTo: targetOne[0].fields[1].name,
        // lineWidth: 0,
        // showPoints: 'always',
        // pointSize: 12,
        hideFrom: {
          "legend": true,
          "viz": true,
          "tooltip": true
        }
      }
    } else {
      item.fields[1].config.custom.showPoints = 'auto'
      item.fields[1].config.custom.pointSize = 4
    }
  })

  return frames;
}

const Index = (props: any) => {
const {
  data,
  timeRange,
  timeZone,
  width,
  height,
  options,
  // fieldConfig,
  // onChangeTimeRange,
  replaceVariables,
  // onOptionsChange
  // id,
} = props;

const [start, setStart] = useState<any>(new Date())
const [end, setEnd] = useState<any>(new Date())
// console.log('=====>DataSourcePicker', DataSourcePicker);
// console.log('=====>config', config);
// console.log('=====>PluginPage', PluginPage);
console.log('=====>pros', props);
useEffect(() => {
  setStart(props.timeRange.from)
  setEnd(props.timeRange.to)
}, [props.timeRange])
// useEffect(() => {
//   const sub = props.eventBus.subscribe(RefreshEvent, () => {
//     console.log('=====>112121212 refresh', );
//   });

//   return () => {
//     sub.unsubscribe();
//   };
// }, [ props.eventBus]);
// const [count, setCount] = useState(0)
const { sync } = usePanelContext();

const theme2 = UI.useTheme2()
// const theme2 = UI.useTheme()
// console.log('=====>theme2',theme2 );
// frames
const frames = useMemo(() => prepareGraphableFields( data.series, theme2, timeRange), [data, timeRange, theme2]) || [];
const timezones = useMemo(() => getTimezones(options.timezone, timeZone), [options.timezone, timeZone]);
// console.log('=====>frames11111', frames);
console.log('=====>timeZone', timezones);
const _cloneframes = pickAndBuild(frames, props.options);
console.log('=====>_cloneframes', _cloneframes);
// props.eventBus.subscribe('dashboard-panels-changed', (evt: any) => {
//   // Remove current selection when entering edit mode for any panel in dashboard
//   // this.scene.clearCurrentSelection();
//   // this.closeInlineEdit();
//   console.log('=====>11111', evt);
// })
// props.eventBus.subscribe('dashboard-meta-changed', (evt: any) => {
//   // Remove current selection when entering edit mode for any panel in dashboard
//   // this.scene.clearCurrentSelection();
//   // this.closeInlineEdit();
//   console.log('=====>2222', evt);
// })
// props.eventBus.getStream('dashboard-meta-changed').subscribe({
//   next: () => {
//     console.log('=====>121212', );
//     // this.doSearch();
//   },
// })

// useEffect(() => {
//   console.log('=====>11mmmmmmm', this);
//   onOptionsChange({c:2})
// }, [])
// DateTimeParser
  return <>
  {/* <div style={{width: '20px', height: '205px'}}></div> */}
  {/* <Refresh callBack={() => {console.log('=====>ssssssss', );}} baseTime={5}/> */}
  {/* <div id='main' style={{width: width, height: height, marginLeft: '20px'}} ref={domRef}></div> */}
  <TimeSeries
      // onClick={()=>{}}
      frames={_cloneframes}
      structureRev={data.structureRev}
      timeRange={{from: dateTimeParse(start), to: dateTimeParse(end), raw: timeRange.raw}}
      timeZone={timezones}
      width={width}
      height={height}
      legend={{
        ...options.legend,
      }}
      options={options}
    >
      {(config, alignedDataFrame) => {
        // console.log('=====>config',config );
        // console.log('=====>alignedDataFrame',alignedDataFrame );
        if (
          alignedDataFrame.fields.filter((f) => f.config.links !== undefined && f.config.links.length > 0).length > 0
        ) {
          alignedDataFrame = regenerateLinksSupplier(alignedDataFrame, _cloneframes, replaceVariables, timeZone);
        }
        return (
          <>
            <KeyboardPlugin config={config} />
            <ZoomPlugin config={config} onZoom={(evt) => {
              setStart(evt.from)
              setEnd(evt.to)
            }} />
            {options?.tooltip?.mode === TooltipDisplayMode.None || (
              <TooltipPlugin
                frames={_cloneframes}
                data={alignedDataFrame}
                config={config}
                mode={TooltipDisplayMode.Multi}
                sortOrder={options.tooltip?.sort}
                sync={sync}
                timeZone={timeZone}
                renderTooltip={(a, ...b) => {
                  const showIndex = b[1];
                  const showLenged = a?.fields.filter((item: any) => {
                    return !item.config.custom?.hideFrom?.viz && item.name !== 'Time' && !/\-anomaly$/g.test(item.name)
                  })
                  const timeX = a?.fields.filter((item: any) => {
                    return item.name === 'Time'
                  })
                  // console.log('=====>showLenged', showLenged);
                  return  showLenged.length ? 
                    <div>
                      {
                        timeX.length ?
                        <div>{showIndex ? dateTimeParse(timeX[0].values.toArray()[showIndex]).format('YYYY-MM-DD HH:mm:ss') : ''}</div>
                        : null
                      }
                      {
                        showLenged.map((item: any, index: number) => {
                         
                          return <div style={{display: 'flex', justifyContent: 'space-between'}} key={index}>
                             <div style={{fontSize: '10', transform: 'scale(.8)', transformOrigin: '0 0'}}>
                            <span style={{display: 'inline-block', width: 15, height: 5, marginRight: '10px', background: item.config.color.fixedColor}}></span>

                            <span>{item.config.displayNameFromDS}</span>
                          </div>
                          <div style={{alignItems: 'flex-end'}}>{showIndex ? item.values.toArray()[showIndex] : ''}</div> 

                          </div>
                          
                         
                        })
                      }
                    </div>
                    : null
                  
                }}
              />
            )}
            {/* <OutsideRangePlugin config={config} onChangeTimeRange={onChangeTimeRange} /> */}
          </>
        );
      }}
      </TimeSeries>
  </>
}
export default Index;
