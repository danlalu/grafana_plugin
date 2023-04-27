import { SimpleOptions } from './types';
import { commonOptionsBuilder } from '@grafana/ui';
import  SimplePanel from './chart';
import {
  FieldColorModeId,
  FieldConfigProperty,
  PanelPlugin
} from '@grafana/data';
import _ from 'lodash'
import { initData } from 'utils';

export const plugin = new PanelPlugin<SimpleOptions>(SimplePanel)
.useFieldConfig({
  disableStandardOptions: [FieldConfigProperty.Color,FieldConfigProperty.Unit, FieldConfigProperty.Decimals, FieldConfigProperty.DisplayName, FieldConfigProperty.NoValue, FieldConfigProperty.Links, FieldConfigProperty.Mappings, FieldConfigProperty.Thresholds, FieldConfigProperty.Min, FieldConfigProperty.Max],
  standardOptions: {
      [FieldConfigProperty.Color]: {
        settings: {
          byValueSupport: false,
          bySeriesSupport: true,
          preferThresholdsMode: false,
        },
        defaultValue: {
          mode: FieldColorModeId.PaletteClassic,
        },
      },
    },
    useCustomConfig: (builder) => {
      commonOptionsBuilder.addHideFrom(builder);
    }
})
.setPanelOptions((builder: any, ...a) => {
  const data = a[0]?.data || [];
  builder
    .addBooleanSwitch({
      path: [`BoundaryWithLenged`],
      name: `Show AI thresholds`,
      defaultValue: true,
    }) 
  initData(data)
  commonOptionsBuilder.addLegendOptions(builder);
  // if(Object.keys(labelKeyMapUid).length) {
  //   Object.keys(labelKeyMapUid).forEach((item, index) => {
  //     builder
  //       .addBooleanSwitch({
  //         path: [`${labelKeyMapUid[item].labelKey}Show`],
  //         // name: `数据列：${item}，是否展示上下界`,
  //         defaultValue: false,
  //         // settings
  //         // category: 'q',
  //         description: `是否展示上下界：${item}`,
  //         showIf: (config: any) => !config.BoundaryWithLenged,
  //       })
  //   })
  // }
})
