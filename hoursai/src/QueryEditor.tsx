import { defaults } from 'lodash';
// import {promLanguageDefinition} from './monaco-promql';
// import {languageConfiguration,language,completionItemProvider} from './promql';
import React, { PureComponent, SyntheticEvent } from 'react';
import { Select, Button, ReactMonacoEditor } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from './datasource';
import { defaultQuery, MyDataSourceOptions, MyQuery } from './types';
import { css } from '@emotion/css';
import {QueryHeaderSwitch} from "./QueryHeaderSwitch"
import { getCompletionProvider } from './completionProvider';
import AlertComponent from "./components/Alert"
// import { EditorField, EditorRow, EditorSwitch } from '@grafana/experimental';
//import MonacoQueryFieldWrapper from "./components/monaco-query-field/MonacoQueryFieldWrapper"
// const { Switch } = LegacyForms;

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;
interface State {
  algorithmList: any;
  expr: any;
  A_Realtime_Save: any;
  errorText: any;
  styles: any;
  placeholder: any;
}
 function limitSuggestions(items: string[]) {
  return items.slice(0, 10000);
}
function processLabels(labels: Array<{ [key: string]: string }>, withName = false) {
  // For processing we are going to use sets as they have significantly better performance than arrays
  // After we process labels, we will convert sets to arrays and return object with label values in arrays
  const valueSet: { [key: string]: Set<string> } = {};
  labels.forEach((label) => {
    const { __name__, ...rest } = label;
    if (withName) {
      valueSet['__name__'] = valueSet['__name__'] || new Set();
      if (!valueSet['__name__'].has(__name__)) {
        valueSet['__name__'].add(__name__);
      }
    }

    Object.keys(rest).forEach((key) => {
      if (!valueSet[key]) {
        valueSet[key] = new Set();
      }
      if (!valueSet[key].has(rest[key])) {
        valueSet[key].add(rest[key]);
      }
    });
  });

  // valueArray that we are going to return in the object
  const valueArray: { [key: string]: string[] } = {};
  limitSuggestions(Object.keys(valueSet)).forEach((key) => {
    valueArray[key] = limitSuggestions(Array.from(valueSet[key]));
  });

  return { values: valueArray, keys: Object.keys(valueArray) };
}
export class QueryEditor extends PureComponent<Props, State> {
  state: State = {
    placeholder: "Enter a PromQL query…",
    algorithmList: [{ label: "Auto", value: "-1" }],
    expr: "",
    A_Realtime_Save: "",
    errorText: "",
    styles: {
      container: css`
        border-radius: 20px;
        border: 1px solid grey;
      `,
      placeholder: css`
        ::after {
          content: 'Enter a PromQL query…';
          opacity: 0.3;
        }
      `,
    }
  };

  onQueryTextChange = (value: any) => {
    // console.error(event,"event")
    // alert(JSON.stringify(event))
    const { onChange, query} = this.props;
    this.setState({ expr: value, errorText: "" })
    onChange({ ...query, expr: value });
    // let params = {
    //   ...query,
    //   ...data?.request,
    //   expr: value,
    //   from: new Date(range?.from as any).getTime() + '',
    //   to: new Date(range?.to as any).getTime() + ''
    // }
    // delete params.range
    // delete params.rangeRaw
    // delete params.targets
    // delete params.scopedVars

    //let algorithmList: any = []
    // this.props.datasource.algorithmList(params,datasource).then((res: any) => {
    //   algorithmList = res?.data || []
    //   let newAlgorithmList: any = algorithmList?.map((item: any) => {
    //     let algorithmListItem = JSON.parse(item)
    //     return {
    //       label: algorithmListItem?.name,
    //       value: algorithmListItem?.id
    //     }
    //   })
    //   this.setState({ algorithmList: [{ label: "Auto", value: "-1" }, ...newAlgorithmList] })
    // })


  };

  onGenericIdChange = (value: any) => {
    const { onChange, query } = this.props;
    onChange({ ...query, genericId: value?.value });
    // executes the query
    //onRunQuery();
  };
  onAlerRuleListChange = (value: any) => {
    const { onChange, query } = this.props;
    onChange({ ...query, alertTemplateId: value?.value });
    // executes the query
    //onRunQuery();
  };
  alertEnableChange = (event: any) => {
    const { onChange, query ,datasource} = this.props;
    onChange({ ...query, alertEnable: event.currentTarget.checked });

    if (event.currentTarget.checked) {
      this.props.datasource.createAlert({ alertTemplateId: query.alertTemplateId, taskId: this.state.A_Realtime_Save },datasource)
    } else {
      this.props.datasource.removeAlert({ taskId: this.state.A_Realtime_Save },datasource)
    }
    // executes the query
    //onRunQuery();
  };
  onRunQueries = () => {
    const { onRunQuery } = this.props;
    // executes the query
    onRunQuery();
  };
  onChangeQuery = () => { }

  realtimeCheckChange = (event: SyntheticEvent<HTMLInputElement>) => {
    const { onChange, query, data ,datasource,range} = this.props;
    console.error(range)
    if (data?.request?.panelId) {

      let params: any = {
        ...query,
        ...data?.request,
      }
      delete params.range
      delete params.rangeRaw
      delete params.targets
      delete params.scopedVars
      if (event.currentTarget.checked) {
       
        this.props.datasource.realtimeTaskSave(params,datasource).then((res: any) => {
          if(res.status!=="error"){
            let A_Realtime_Save: any = ""
            try {
              A_Realtime_Save = res?.data?.taskId
            } catch {
              A_Realtime_Save = ""
            }
            this.setState({ A_Realtime_Save: A_Realtime_Save })
            
            onChange({ ...query, queryType: "realtimeCheck", A_Realtime_Save: A_Realtime_Save });
          }else{
          //  event.currentTarget.checked=false
            onChange({ ...query, queryType: "syncPreview" });
          }
          
        })
      } else {
        params = {
          ...query,
          taskId: this.state.A_Realtime_Save
        }
          
          this.props.datasource.realtimeTaskRemove(params,datasource).then((res: any)=>{
            if(res.status!=="error"){
              this.setState({ A_Realtime_Save: "" })
              onChange({ ...query, queryType: "syncPreview", A_Realtime_Save: "" });
            }else{
             // event.currentTarget.checked=true
              onChange({ ...query, queryType: "realtimeCheck" });
            }
        })
       
       
       
      }
    } else {
      this.setState({ errorText: "You can associate only after saving. Please save first!" })
    }
    // this.setState({A_Realtime_Save:"1"})

  };
  componentDidMount() {
    // alert(JSON.stringify(window.location.href))
    const { data, range, datasource, query } = this.props;
    if (query?.expr) {
      this.setState({ expr: query?.expr })
    }
    if (query?.A_Realtime_Save) {
      this.setState({ A_Realtime_Save: query?.A_Realtime_Save })
    }
   
   
    let params = {
      ...query,
      ...data?.request,
      from: new Date(range?.from as any).getTime() + '',
      to: new Date(range?.to as any).getTime() + ''
    }
    delete params.range
    delete params.rangeRaw
    delete params.targets
    delete params.scopedVars
    // this.setState({algorithmList:[
    //   { label: 'Time series', value: 'time_series' },
    //   { label: 'Table', value: 'table' },
    //   { label: 'Heatmap', value: 'heatmap' },
    // ]})
    let algorithmList: any = []
    this.props.datasource.algorithmList(params,datasource).then((res: any) => {
      algorithmList = res?.data || []
      let newAlgorithmList: any = algorithmList?.map((item: any) => {
        let algorithmListItem = JSON.parse(item)
        return {
          label: algorithmListItem?.name,
          value: algorithmListItem?.id
        }
      })
      this.setState({ algorithmList: [{ label: "Auto", value: "-1" }, ...newAlgorithmList] })
    })

  }

  editorWillMount(monaco: any) {


    const languageConfiguration: any = {
      // the default separators except `@$`
      wordPattern: /(-?\d*\.\d\w*)|([^`~!#%^&*()\-=+\[{\]}\\|;:'",.<>\/?\s]+)/g,
      // Not possible to make comments in PromQL syntax
      comments: {
        lineComment: '#',
      },
      brackets: [
        ['{', '}'],
        ['[', ']'],
        ['(', ')'],
      ],
      autoClosingPairs: [
        { open: '{', close: '}' },
        { open: '[', close: ']' },
        { open: '(', close: ')' },
        { open: '"', close: '"' },
        { open: '\'', close: '\'' },
      ],
      surroundingPairs: [
        { open: '{', close: '}' },
        { open: '[', close: ']' },
        { open: '(', close: ')' },
        { open: '"', close: '"' },
        { open: '\'', close: '\'' },
        { open: '<', close: '>' },
      ],
      folding: {}
    };
    // PromQL Aggregation Operators
    // (https://prometheus.io/docs/prometheus/latest/querying/operators/#aggregation-operators)
    const aggregations: any = [
      'sum',
      'min',
      'max',
      'avg',
      'group',
      'stddev',
      'stdvar',
      'count',
      'count_values',
      'bottomk',
      'topk',
      'quantile',
    ];
    // PromQL functions
    // (https://prometheus.io/docs/prometheus/latest/querying/functions/)
    const functions: any = [
      'abs',
      'absent',
      'ceil',
      'changes',
      'clamp_max',
      'clamp_min',
      'day_of_month',
      'day_of_week',
      'days_in_month',
      'delta',
      'deriv',
      'exp',
      'floor',
      'histogram_quantile',
      'holt_winters',
      'hour',
      'idelta',
      'increase',
      'irate',
      'label_join',
      'label_replace',
      'ln',
      'log2',
      'log10',
      'minute',
      'month',
      'predict_linear',
      'rate',
      'resets',
      'round',
      'scalar',
      'sort',
      'sort_desc',
      'sqrt',
      'time',
      'timestamp',
      'vector',
      'year',
    ];

    const aggregationsOverTime: any = [];
    for (let _i = 0, aggregations_1 = aggregations; _i < aggregations_1.length; _i++) {
      const agg = aggregations_1[_i];
      aggregationsOverTime.push(agg + '_over_time');
    }

    const vectorMatching = [
      'on',
      'ignoring',
      'group_right',
      'group_left',
      'by',
      'without',
    ];
    // Produce a regex matching elements : (elt1|elt2|...)
    const vectorMatchingRegex = "(" + vectorMatching.reduce(function (prev, curr) { return prev + "|" + curr; }) + ")";
    // PromQL Operators
    // (https://prometheus.io/docs/prometheus/latest/querying/operators/)
    const operators = [
      '+', '-', '*', '/', '%', '^',
      '==', '!=', '>', '<', '>=', '<=',
      'and', 'or', 'unless',
    ];
    // PromQL offset modifier
    // (https://prometheus.io/docs/prometheus/latest/querying/basics/#offset-modifier)
    const offsetModifier = [
      'offset',
    ];
    // Merging all the keywords in one list
    const keywords = aggregations.concat(functions).concat(aggregationsOverTime).concat(vectorMatching).concat(offsetModifier);
    // noinspection JSUnusedGlobalSymbols
    const language = {
      ignoreCase: false,
      defaultToken: '',
      tokenPostfix: '.promql',
      keywords: keywords,
      operators: operators,
      vectorMatching: vectorMatchingRegex,
      // we include these common regular expressions
      symbols: /[=><!~?:&|+\-*\/^%]+/,
      escapes: /\\(?:[abfnrtv\\"']|x[0-9A-Fa-f]{1,4}|u[0-9A-Fa-f]{4}|U[0-9A-Fa-f]{8})/,
      digits: /\d+(_+\d+)*/,
      octaldigits: /[0-7]+(_+[0-7]+)*/,
      binarydigits: /[0-1]+(_+[0-1]+)*/,
      hexdigits: /[[0-9a-fA-F]+(_+[0-9a-fA-F]+)*/,
      integersuffix: /(ll|LL|u|U|l|L)?(ll|LL|u|U|l|L)?/,
      floatsuffix: /[fFlL]?/,
      // The main tokenizer for our languages
      tokenizer: {
        root: [
          // 'by', 'without' and vector matching
          [/@vectorMatching\s*(?=\()/, 'type', '@clauses'],
          // labels
          [/[a-z_]\w*(?=\s*(=|!=|=~|!~))/, 'tag'],
          // comments
          [/(^#.*$)/, 'comment'],
          // all keywords have the same color
          [/[a-zA-Z_]\w*/, {
            cases: {
              '@keywords': 'type',
              '@default': 'identifier'
            }
          }],
          // strings
          [/"([^"\\]|\\.)*$/, 'string.invalid'],
          [/'([^'\\]|\\.)*$/, 'string.invalid'],
          [/"/, 'string', '@string_double'],
          [/'/, 'string', '@string_single'],
          [/`/, 'string', '@string_backtick'],
          // whitespace
          { include: '@whitespace' },
          // delimiters and operators
          [/[{}()\[\]]/, '@brackets'],
          [/[<>](?!@symbols)/, '@brackets'],
          [/@symbols/, {
            cases: {
              '@operators': 'delimiter',
              '@default': ''
            }
          }],
          // numbers
          [/\d+[smhdwy]/, 'number'],
          [/\d*\d+[eE]([\-+]?\d+)?(@floatsuffix)/, 'number.float'],
          [/\d*\.\d+([eE][\-+]?\d+)?(@floatsuffix)/, 'number.float'],
          [/0[xX][0-9a-fA-F']*[0-9a-fA-F](@integersuffix)/, 'number.hex'],
          [/0[0-7']*[0-7](@integersuffix)/, 'number.octal'],
          [/0[bB][0-1']*[0-1](@integersuffix)/, 'number.binary'],
          [/\d[\d']*\d(@integersuffix)/, 'number'],
          [/\d(@integersuffix)/, 'number'],
        ],
        string_double: [
          [/[^\\"]+/, 'string'],
          [/@escapes/, 'string.escape'],
          [/\\./, 'string.escape.invalid'],
          [/"/, 'string', '@pop']
        ],
        string_single: [
          [/[^\\']+/, 'string'],
          [/@escapes/, 'string.escape'],
          [/\\./, 'string.escape.invalid'],
          [/'/, 'string', '@pop']
        ],
        string_backtick: [
          [/[^\\`$]+/, 'string'],
          [/@escapes/, 'string.escape'],
          [/\\./, 'string.escape.invalid'],
          [/`/, 'string', '@pop']
        ],
        clauses: [
          [/[^(,)]/, 'tag'],
          [/\)/, 'identifier', '@pop']
        ],
        whitespace: [
          [/[ \t\r\n]+/, 'white'],
        ],
      },
    };
    // noinspection JSUnusedGlobalSymbols
    const completionItemProvider = {
      provideCompletionItems: function () {
        // To simplify, we made the choice to never create automatically the parenthesis behind keywords
        // It is because in PromQL, some keywords need parenthesis behind, some don't, some can have but it's optional.
        const suggestions = keywords.map(function (value: any) {
          return {
            label: value,
            kind: monaco.languages.CompletionItemKind.Keyword,
            insertText: value,
            insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet
          };
        });
        return { suggestions: suggestions };
      }
    };



    const languageId = "promql";
    monaco.languages.register({
      id: 'promql',
      extensions: ['.promql'],
      aliases: ['Prometheus', 'prometheus', 'prom', 'Prom', 'promql', 'Promql', 'promQL', 'PromQL'],
      mimetypes: [],
      //loader: function () { return import('./monaco-promql/promql/promql'); } // eslint-disable-line @typescript-eslint/explicit-function-return-type
    });
    monaco.languages.setMonarchTokensProvider(languageId, language);
    monaco.languages.setLanguageConfiguration(languageId, languageConfiguration);
    monaco.languages.registerCompletionItemProvider(languageId, completionItemProvider);
  }
  // fetchLabelValues = async (key: string): Promise<string[]> => {
  //   const params = this.datasource.getTimeRangeParams();
  //   const url = `/api/v1/label/${this.datasource.interpolateString(key)}/values`;
  //   return await this.request(url, [], params);
  // };

  // async getLabelValues(key: string): Promise<string[]> {
  //   return await this.fetchLabelValues(key);
  // }
  render() {
    const option: any = {
      codeLens: false,
      contextmenu: false,
      // we need `fixedOverflowWidgets` because otherwise in grafana-dashboards
      // the popup is clipped by the panel-visualizations.
      fixedOverflowWidgets: true,
      folding: false,
      fontSize: 14,
      lineDecorationsWidth: 8, // used as "padding-left"
      lineNumbers: 'off',
      minimap: { enabled: false },
      overviewRulerBorder: false,
      overviewRulerLanes: 0,
      padding: {
        // these numbers were picked so that visually this matches the previous version
        // of the query-editor the best
        top: 4,
        bottom: 5,
      },
      renderLineHighlight: 'none',
      scrollbar: {
        vertical: 'hidden',
        verticalScrollbarSize: 8, // used as "padding-right"
        horizontal: 'hidden',
        horizontalScrollbarSize: 0,
      },
      scrollBeyondLastLine: false,
      suggest: { showWords: false },
      suggestFontSize: 12,
      wordWrap: 'on',
    };
    const query = defaults(this.props.query, defaultQuery);
    const { expr, genericId = "", queryType} = query;
    // const alerRuleList: any = [
    //   { label: 'the value is ABOVE upper threshold', value: 1 },
    //   { label: 'the value is BELOW lower threshold', value: 2 },
    //   { label: 'the value is OUTSIDE of the range', value: 3 },
    // ];
    return (<>
       {window.location.href.indexOf("/alerting")!==-1&&<div style={{ width: "100%" }}>
          <AlertComponent datasource={this.props.datasource} query={query} onChange={this.props.onChange} runQuery={this.props.onRunQuery}></AlertComponent>
        </div>}
      {window.location.href.indexOf("/alerting")===-1&&<div style={{ width: "100%" }}>
        <div style={{display:"flex", justifyContent: "flex-end",margin: "10px",marginRight: "40px"}}>
          <Button 
          disabled={!this.state.expr}
          size="sm"
          type="button" 
          icon="sync"
           onClick={this.onRunQueries}>
            Run queries
          </Button>
        </div>
         
        <div style={{ justifyContent: "space-between", width: "100%" }} className="gf-form">

          {/* <FormField
              labelWidth={10}
              value={expr || ''}
              onChange={this.onQueryTextChange}
              label="Expression"
              tooltip="Expression"
              inputWidth={350}
              style={{width:"500px"}}
            /> */}

          <div style={{width:"100%"}} className="gf-form gf-form-switch-container">
            <div className='gf-form-label'>Query</div>
            <ReactMonacoEditor
              className="ReactMonacoEditor"
              options={option}
              language="promql"
              value={expr || ''}
              beforeMount={this.editorWillMount}
              onChange={this.onQueryTextChange}
              onMount={(editor, monaco) => {
                const {datasource,range} = this.props;
                // editor.onDidBlurEditorWidget(() => {
                //   onBlurRef.current(editor.getValue());
                // });

               // const getSeries = (selector: string) => lpRef.current.getSeries(selector);
               const getSeries = async (selector: string) =>{
                // let data= [{
                //   "__name__": "up",
                //   "instance": "10.1.23.223:9090",
                //   "job": "prometheus"}
                // ]
                let start: any=range?.from.utc().format('YYYY-MM-DD HH:mm');
                let end: any=range?.to.utc().format('YYYY-MM-DD HH:mm');
                const metrics: any= await this.props.datasource.getSeries({start:new Date(start).getTime(),end:new Date(end).getTime(),"match[]":selector},datasource)  

                   const { values } = processLabels(metrics?.data)
                   return Promise.resolve(values)

               }

                // const getHistory = () =>
                //   Promise.resolve(historyRef.current.map((h) => h.query.expr).filter((expr) => expr !== undefined));
               
                const getAllMetricNames = async () => {
                  let start: any=range?.from.utc().format('YYYY-MM-DD HH:mm')
                  let end: any=range?.to.utc().format('YYYY-MM-DD HH:mm')
                  const metrics: any= await this.props.datasource.getMetrics({start:new Date(start).getTime(),end:new Date(end).getTime()},datasource)  
                  const result = metrics?.data?.map((m: any) => {
                 
                    return {
                      name: m,
                      help: '',
                      type: '',
                    };
                  });
                  return Promise.resolve(result);
                };
                // const getAllMetricNames = () => {
                //   const metrics = [
                //     "scrape_duration_seconds",
                //     "scrape_samples_post_metric_relabeling",
                //     "scrape_samples_scraped",
                //     "scrape_series_added",
                //     "up"
                // ]
                //   const result = metrics.map((m: any) => {
                 
                //     return {
                //       name: m,
                //       help: '',
                //       type: '',
                //     };
                //   });

                //   return Promise.resolve(result);
                // };

               
                 //const getAllLabelNames = () => Promise.resolve(lpRef.current.getLabelKeys());
                 const getAllLabelNames = async() => {
                  let start: any=range?.from.utc().format('YYYY-MM-DD HH:mm')
                  let end: any=range?.to.utc().format('YYYY-MM-DD HH:mm')
                  const metrics: any= await this.props.datasource.getLabelNames({start:new Date(start).getTime(),end:new Date(end).getTime()},datasource)  

                  return Promise.resolve(metrics.data.slice().sort());
                 
                };

                // const getLabelValues = (labelName: string) => lpRef.current.getLabelValues(labelName);
                //const getLabelValues = (labelName: string) => lpRef.current.getLabelValues(labelName);

                const dataProvider = {getSeries, getAllMetricNames,getAllLabelNames};
                const completionProvider = getCompletionProvider(monaco, dataProvider);

                const filteringCompletionProvider: any = {
                  ...completionProvider,
                  provideCompletionItems: (model: any, position: any, context: any, token: any) => {
                    // if (editor.getModel()?.id !== model.id) {
                    //   return { suggestions: [] };
                    // }
                    return completionProvider.provideCompletionItems(model, position, context, token);
                  },
                };

                // const { dispose } =
                 monaco.languages.registerCompletionItemProvider(
                  "promql",
                  filteringCompletionProvider
                );

                //autocompleteDisposeFun.current = dispose;
                const updateElementHeight = () => {
                  const containerDiv: any = document.querySelector(".ReactMonacoEditor");
                  if (containerDiv !== null) {
                    const pixelHeight = editor.getContentHeight();
                    containerDiv.style.height = `${pixelHeight + 2}px`;
                    containerDiv.style.width = 'calc(100%-20px)';
                    const pixelWidth = containerDiv.clientWidth;
                    editor.layout({ width: pixelWidth, height: pixelHeight });
                  }
                };

                editor.onDidContentSizeChange(updateElementHeight);
                updateElementHeight();
                editor.getModel()?.onDidChangeContent(() => {
                  // onChangeRef.current(editor.getValue());
                });
                // editor.addCommand(monaco.KeyMod.Shift | monaco.KeyCode.Enter, () => {
                //   onRunQueryRef.current(editor.getValue());
                // });

                // editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyK, function () {
                //   global.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', metaKey: true }));
                // });

                if (this.state.placeholder) {
                  const placeholderDecorators = [
                    {
                      range: new monaco.Range(1, 1, 1, 1),
                      options: {
                        className: this.state.styles.placeholder,
                        isWholeLine: true,
                      },
                    },
                  ];

                  let decorators: string[] = [];

                  const checkDecorators: () => void = () => {
                    const model = editor.getModel();

                    if (!model) {
                      return;
                    }

                    const newDecorators = model.getValueLength() === 0 ? placeholderDecorators : [];
                    decorators = model.deltaDecorations(decorators, newDecorators);
                  };

                  checkDecorators();
                  editor.onDidChangeModelContent(checkDecorators);
                }
              }}

            />
          </div>
         
        </div>
        <div className="gf-form gf-form-switch-container">
          <div className='gf-form-label'>AI model</div>
          <Select   value={genericId} placeholder={"algorithm"} allowCustomValue onChange={this.onGenericIdChange} options={this.state.algorithmList} />

          </div>
        <div className="gf-form">


          <QueryHeaderSwitch  disabled={!this.props.data?.request?.panelId}   value={queryType === "realtimeCheck" ? true : false} label="Start real time monitoring" onChange={this.realtimeCheckChange} />
        
        </div>
  
        {/* {
            queryType === "realtimeCheck" ?
              <div style={{alignItems: "center"}} className="gf-form">
                // <FormField
                //     labelWidth={0}
                //     value={this.state.A_Realtime_Save || ''}
                //     label="Task ID"
                //     tooltip="Task ID"
                //     disabled={true}
                //     inputWidth={0}
                //   />   
                <div style={{width:"200px"}}>
                 <QueryHeaderSwitch value={alertEnable} label="Set Alert on Condition" onChange={this.alertEnableChange} />

                </div>
                <Select  width={100} value={alertTemplateId} placeholder={"Alert Operations"} allowCustomValue onChange={this.onAlerRuleListChange} options={alerRuleList} />
              </div>

              : null
          } */}

        <div className="gf-form">
          <span style={{ color: 'grey', lineHeight: '30px' }}>
            {this.state.errorText}
          </span>
        </div>


      </div>}
      </>

    );
  }
}
