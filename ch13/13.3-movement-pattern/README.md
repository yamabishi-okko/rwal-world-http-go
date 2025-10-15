観察ポイント（Networkタブ
Go SSR: 遷移/再読込ごとに document (HTML) が返る<br>
Go SSR+Ajax: 初回は document、ボタン以降は /api/state の JSON<br>
React SPA: 最初に index.html と app のJS、以降はクライアント内で更新（今回は擬似）<br>
Next SSR+SPA: 初回は SSR された document + JS、以降は /api/state の JSON<br>