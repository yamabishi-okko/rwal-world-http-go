import React, { useEffect, useState } from 'react'


export default function App(){
const [now, setNow] = useState(new Date().toISOString())
const [items, setItems] = useState(["apple","banana","cherry"]) // 擬似API初期値


async function refresh(){
// 実APIがない前提の最小:
setNow(new Date().toISOString())
// items も擬似更新
setItems(["apple","banana","cherry"].map(x=>x+"*"))
}


useEffect(()=>{ /* 初回に擬似フェッチ */ },[])


return (
<main>
<h1>React SPA</h1>
<p>Client Time: {now}</p>
<ul>{items.map(x=> <li key={x}>{x}</li>)}</ul>
<button onClick={refresh}>時刻とitemsを更新（擬似）</button>
</main>
)
}