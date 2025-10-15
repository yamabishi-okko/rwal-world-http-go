export async function getServerSideProps(){
    // SSRで初期データを作る（本来はDB/APIを叩く）
    return {
    props: {
    initial: { now: new Date().toISOString(), items: ["apple","banana","cherry"] }
    }
    }
    }
    
    
    export default function Home({ initial }){
    async function refresh(){
    const r = await fetch('/api/state')
    const j = await r.json()
    document.getElementById('now').textContent = j.now
    document.getElementById('items').innerHTML = j.items.map(x=>`<li>${x}</li>`).join('')
    }
    return (
    <main>
    <h1>Next SSR + SPA</h1>
    <p>Initial (SSR) Time: {initial.now}</p>
    <p>Live Time: <span id="now">{initial.now}</span></p>
    <ul id="items">{initial.items.map(x=> <li key={x}>{x}</li>)}</ul>
    <button onClick={refresh}>APIで更新</button>
    </main>
    )
}