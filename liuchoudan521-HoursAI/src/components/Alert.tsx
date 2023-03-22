import React, {useEffect, useState } from 'react';
import { Select, Button } from '@grafana/ui';
import { getBackendSrv } from '@grafana/runtime';
export default function Alert(props: any) {
    const { query, onChange, runQuery, datasource } = props
    const [taskIdList, settaskIdList] = useState<any>([]);
    const [exprList, setExprList] = useState<any>([]);
    const [merticsList] = useState<any>([
         { label: "upper", value: "upper"},
         { label: "lower", value: "lower"},
         { label: "baseline", value: "baseline"},
         { label: "anomaly", value: "anomaly"},
         { label: "significance", value: "significance"}]);
    const [taskId, setTaskId] = useState<any>(query?.A_Realtime_Save||"")
    const [mergicId, setMertic] = useState<any>(query?.series||"")
    const [label, setLabel] = useState<any>("TaskId")
    const [labelList] = useState<any>([{label: "TaskId", value: "TaskId"},{ label: "Expr", value: "Expr"}]);
   
    const taskIdHandelChange = (e: any) => {
        setTaskId(e.value)
        onChange({ ...query, A_Realtime_Save: e.value,queryType: "realtimeResultSingle" })
    }
    const merticHandelChange = (e: any) => {
        setMertic(e.value)
        onChange({ ...query, series: e.value,queryType: "realtimeResultSingle"  })
    }
    const labelHandelChange = (e: any) => {
        setLabel(e.value)
       
    }
    useEffect(()=>{
        getBackendSrv().post(`/api/datasources/${datasource.id}/resources/realtimeTaskList`).then((res: any) => {
            const result: any = res?.data||[]
            const tasks: any = result?.map((item: any) => {
                return { label:`${datasource?.name||""}-${item.taskId}` , value: item.taskId }
            })
            settaskIdList(tasks)
            const exprs: any = result?.map((item: any) => {
                return { label: item.query, value: item.taskId }
            })
            setExprList(exprs)
        })
    },[datasource])
    return <>
        <div style={{ display: "flex", justifyContent: "flex-end", margin: "10px", marginRight: "40px" }}>
            <Button
                disabled={!taskId || !mergicId}
                size="sm"
                type="button"
                icon="sync"
                onClick={runQuery}>
                Run queries
            </Button>
        </div>
        <div className="gf-form gf-form-switch-container">
            <div className='gf-form-label' style={{padding:0}}>
                <Select value={label} placeholder={"taskId"} allowCustomValue onChange={labelHandelChange} options={labelList} />
            </div>
            {label === "TaskId" ?
                <Select value={taskId} placeholder={"TaskId"} onChange={taskIdHandelChange} options={taskIdList} /> :
                <Select value={taskId} placeholder={"Exer"} onChange={taskIdHandelChange} options={exprList} />
            }

        </div>
        <div className="gf-form gf-form-switch-container">
            <div className='gf-form-label'>Series</div>
            <Select value={mergicId} placeholder={"Series"} allowCustomValue onChange={merticHandelChange} options={merticsList} />
        </div>
    </>
};
