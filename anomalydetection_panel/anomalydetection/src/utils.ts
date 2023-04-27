import {
  ArrayVector,
  DataFrame,
  Field,
  FieldType,
  getDisplayProcessor,
  getLinksSupplier,
  GrafanaTheme2,
  InterpolateFunction,
  isBooleanUnit,
  SortedVector,
  TimeRange,
} from '@grafana/data';
import { GraphFieldConfig, LineInterpolation } from '@grafana/schema';
import { applyNullInsertThreshold } from './nullInsertThreshold';
import { nullToValue } from './nullToValue';
import _ from 'lodash'

// import {GraphNG} from '@grafana/ui';
/**
 * Returns null if there are no graphable fields
 */
export function prepareGraphableFields(
  series: DataFrame[],
  theme: GrafanaTheme2,
  timeRange?: TimeRange
): DataFrame[] | null {
  if (!series?.length) {
    return null;
  }

  let copy: Field;

  const frames: DataFrame[] = [];

  for (let frame of series) {
    const fields: Field[] = [];

    let hasTimeField = false;
    let hasValueField = false;

    let nulledFrame = applyNullInsertThreshold({
      frame,
      refFieldPseudoMin: timeRange?.from.valueOf(),
      refFieldPseudoMax: timeRange?.to.valueOf(),
    });

    for (const field of nullToValue(nulledFrame).fields) {
      switch (field.type) {
        // console.log('=====>', );
        case FieldType.time:
          hasTimeField = true;
          fields.push(field);
          break;
        case FieldType.number:
          hasValueField = true;
          copy = {
            ...field,
            values: new ArrayVector(
              field.values.toArray().map((v) => {
                if (!(Number.isFinite(v) || v == null)) {
                  return null;
                }
                return v;
              })
            ),
          };

          fields.push(copy);
          break; // ok
        case FieldType.string:
          copy = {
            ...field,
            values: new ArrayVector(field.values.toArray()),
          };

          fields.push(copy);
          break; // ok
        case FieldType.boolean:
          hasValueField = true;
          const custom: GraphFieldConfig = field.config?.custom ?? {};
          const config = {
            ...field.config,
            max: 1,
            min: 0,
            custom,
          };

          // smooth and linear do not make sense
          if (custom.lineInterpolation !== LineInterpolation.StepBefore) {
            custom.lineInterpolation = LineInterpolation.StepAfter;
          }

          copy = {
            ...field,
            config,
            type: FieldType.number,
            values: new ArrayVector(
              field.values.toArray().map((v) => {
                if (v == null) {
                  return v;
                }
                return Boolean(v) ? 1 : 0;
              })
            ),
          };

          if (!isBooleanUnit(config.unit)) {
            config.unit = 'bool';
            copy.display = getDisplayProcessor({ field: copy, theme });
          }

          fields.push(copy);
          break;
      }
    }
    if (hasTimeField && hasValueField) {
      frames.push({
        ...frame,
        // length: nulledFrame.length,
        fields,
      });
    }
  }

  if (frames.length) {
    return frames;
  }

  return null;
}

export function getTimezones(timezones: string[] | undefined, defaultTimezone: string): string[] {
  if (!timezones || !timezones.length) {
    return [defaultTimezone];
  }
  return timezones.map((v) => (v?.length ? v : defaultTimezone));
}

export function regenerateLinksSupplier(
  alignedDataFrame: DataFrame,
  frames: DataFrame[],
  replaceVariables: InterpolateFunction,
  timeZone: string
): DataFrame {
  alignedDataFrame.fields.forEach((field) => {
    const frameIndex = field.state?.origin?.frameIndex;

    if (frameIndex === undefined) {
      return;
    }

    const frame = frames[frameIndex];
    const tempFields: Field[] = [];

    /* check if field has sortedVector values
      if it does, sort all string fields in the original frame by the order array already used for the field
      otherwise just attach the fields to the temporary frame used to get the links
    */
    for (const frameField of frame.fields) {
      if (frameField.type === FieldType.string) {
        if (field.values instanceof SortedVector) {
          const copiedField = { ...frameField };
          copiedField.values = new SortedVector(frameField.values, field.values?.toArray());
          tempFields.push(copiedField);
        } else {
          tempFields.push(frameField);
        }
      }
    }

    const tempFrame: DataFrame = {
      fields: [...alignedDataFrame.fields, ...tempFields],
      length: alignedDataFrame.fields.length + tempFields.length,
    };

    field.getLinks = getLinksSupplier(tempFrame, field, field.state!.scopedVars!, replaceVariables, timeZone);
  });

  return alignedDataFrame;
}

const formatNum = (num: number) => {
  if(num % 1 === 0) {
    return num
  } else {
    const arr: string[] = `${num}`.split('.');
    const target = arr[1];
    const index = target.split('').findIndex((i: string) => i !== '0')
    if(target.length > index + 3) {
      return +Number(`${arr[0]}.${target.substring(0, index + 4)}`).toFixed(index + 3)
    }
    return +num;
  }
}

export function initData (data: any[]) {
  const allFormated = data.every((item: any) => item.hasFormat)
  !allFormated && data.forEach((item: any) => {
    item.hasFormat = true
    const target = item.fields[1]?.values?.toArray();
    if(target) {
      const len = target.length;
      for(let i = 0; i < len; i++) {
        target[i] = formatNum(target[i])
      }
    }
  })
  const KEY_MAP: any = {
    upper: '-Upper',
    lower: '-Lower',
    baseline: '-baseline',
    anomaly: '-anomaly',
    significance: '-significance',
  }
  const COLOR_MAP = [
    {
      main: 'rgb(61, 113, 217)',
      mark: 'rgba(61, 113, 217, .5)',
    },
    {
      main: '#D2691E',
      mark: '#FFDAB9',
    },
    {
      main: '#556B2F',
      mark: '#FAFAD2',
    },
    {
     main: '#00FFFF',
      mark: '#008B8B',
    },
    {
      main: '#FF00FF',
      mark: '#EE82EE',
    },
    {
      main: '#9400D3',
      mark: '#9370DB',
    },
  ];
  const baseNum = COLOR_MAP.length;
  let labelKeyMapUid: any = {};
  let _index = 0;
  
  const expArr = ['upper', 'lower', 'baseline', 'anomaly', 'significance']
  data.forEach((item: any) => {
    // const 
    expArr.forEach((exp: any) => {
      const reg = new RegExp(`^${exp}`)
      if(reg.test(item.name)) {
        const name = JSON.parse(JSON.stringify(item.name))
        item.isReal = name.startsWith(`${exp}_`)
        item.originName = exp
      }
    })
   
    delete item?.fields[1]?.labels.__name__
    item.labelKey = item?.fields[1]?.labels ? item.refId + JSON.stringify(item?.fields[1]?.labels) : '';

    if(!item.uuid) {
      item.uuid = _.uniqueId();
      // item.name = item.refId;
    }
    if(!labelKeyMapUid[item.labelKey]) {
      // labelKeyMapUid[item.labelKey] = item.uuid;
      labelKeyMapUid[item.labelKey] = {
        uuid: item.uuid,
        colorConfig: COLOR_MAP[_index % baseNum],
        labelKey: item.labelKey
      }
      item.showedColor = COLOR_MAP[_index % baseNum].main;
      item.showedColor = COLOR_MAP[_index % baseNum].main
      _index++;
    } else {
      item.showedColor = labelKeyMapUid[item.labelKey].colorConfig.mark

      item.originUuid = labelKeyMapUid[item.labelKey].uuid
      // console.log('=====>12121212', item?.fields[1]?.state?.displayName);
      expArr.forEach((exp: any) => {
        const reg = new RegExp(`^${exp}`)
        if(reg.test(item.name)) {
          item.uuid = labelKeyMapUid[item.labelKey].uuid + KEY_MAP[exp]
          item.fields[1].name = labelKeyMapUid[item.labelKey].uuid + KEY_MAP[exp]
          // item.fields[1].labels = {}
          item.name = `${exp}-${item.labelKey}`
        }
      })
    }
  })
  const anomalyReg = new RegExp(`^anomaly`)

  // if(anomalyList.length) {
  //   anomalyList
  // }
  
  // 处理颜色
  data.forEach((item: any) => {
    if(anomalyReg.test(item.name)) {
      const target = data.filter((source: any) => source.uuid === item.originUuid)
      const times = item.fields[0].values.toArray();

      const value = item.fields[1].values.toArray();
      let indexArr: any = []
      // console.log('=====>value',value );
      value.forEach((v: any, vi: number) => {
        if(v === 0 || v === null) {
          value[vi] = null
        } else {
          indexArr.push(vi)
        }
        // v === 1 && 
      })
      item.fields[1].values = new ArrayVector(value) 
      // console.log('=====>indexArr', indexArr);
      if(indexArr.length) {
        // 拿到所有异常点的时间戳
        const targetTime = target[0].fields[0].values.toArray();
        const targetValue = target[0].fields[1].values.toArray();
        const timeArr = indexArr.map((item: any) => times[item])
        const allIncludes = timeArr.every((item: any) => targetTime.includes(item))
        // console.log('=====>timeArr', timeArr);
        // console.log('=====>allIncludes', allIncludes);
        if(allIncludes) {
          let indexs: any = [];
          targetTime.forEach((item: any, index: number) => {
            if(timeArr.includes(item)) {
              indexs.push(index)
            }
          })
          item.fields[0].values = new ArrayVector(targetTime)
          const result = targetValue.map((item: any, index: number) => {
            return indexs.includes(index) ? item : null
          })
          item.fields[1].values = new ArrayVector(result) 
          // console.log('=====>inde', indexs);
        }
      

      }
      // console.log('=====>indexarr', indexArr);
    }
    item.fields[1].config.color = {
       mode: 'fixed',
       // fixedColor: '#E0ECFF'
       fixedColor: item.showedColor
     }
 })
 data.sort((a, b) => {
  const a_index = getUuid(a.uuid)
  const b_index = getUuid(b.uuid)
  return +a_index - +b_index
 })
 console.log('=====>ddddd', data);
 console.log('=====>labelKeyMapUid', labelKeyMapUid);
//  labelKeyMapUid
  return {labelKeyMapUid, data}
}

const getUuid = (str: any) => {
  let _uuid = ''
  const expArr = ['Upper', 'Lower', 'baseline', 'anomaly', 'significance']
  expArr.forEach((exp: any) => {
    const reg = new RegExp(`\-${exp}$`)
    if(reg.test(str)) {
      const match = str.match(reg)
      if(match) {
        _uuid = str.substr(0, match.index)
      }
    }
  })
  return _uuid === '' ? str : _uuid
}
