const $ = id => document.getElementById(id);
let game, history = [], score, busy = false;
const color = n => n === 1 ? 'Black' : 'White';
async function api(url, body) {
  const res = await fetch(url, body ? {method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)} : {});
  const data = await res.json(); if (!res.ok) throw Error(data.error || 'Request failed'); return data;
}
function star(r,c,n) {
  const a = n === 9 ? [2,4,6] : [3,(n-1)/2,n-4];
  return n === 9 ? (r===4&&c===4)||([2,6].includes(r)&&[2,6].includes(c)) : a.includes(r)&&a.includes(c);
}
function render() {
  const n=game.size, ended=game.passes>=2, board=$('board'), focused=document.activeElement?.dataset?.point;
  board.style.setProperty('--n',n); board.replaceChildren();
  game.board.forEach((v,i)=>{
    const r=Math.floor(i/n), c=i%n, b=document.createElement('button');
    b.type='button'; b.dataset.point=i; b.className=['point',!v?'empty':'',c===0?'left':'',c===n-1?'right':'',r===0?'top-edge':'',r===n-1?'bottom':'',i===game.last?'last':''].filter(Boolean).join(' ');
    b.setAttribute('aria-label',`${'ABCDEFGHJKLMNOPQRST'[c]}${n-r}, ${v?color(v):'empty'}`);
    if(star(r,c,n)){const dot=document.createElement('span');dot.className='star';b.append(dot)}
    const stone=document.createElement('span');stone.className=`stone ${color(v||game.turn).toLowerCase()} ${v?'':'ghost'}`;b.append(stone);
    b.disabled=ended||busy; b.addEventListener('click',()=>move('play',i));board.append(b);
  });
  if(focused!==undefined&&!ended&&!busy) board.querySelector(`[data-point="${focused}"]`)?.focus({preventScroll:true});
  $('status').textContent=ended?(score.black===score.white?'Draw':`${score.black>score.white?'Black':'White'} wins by ${Math.abs(score.black-score.white)}`):`${color(game.turn)} to play${game.passes?' · 1 pass':''}`;
  for(const v of [1,2]){const name=color(v).toLowerCase();$(`${name}-captures`).textContent=game.captures[v];$(`${name}-player`).classList.toggle('active',!ended&&v===game.turn)}
  $('pass').disabled=ended||busy;$('undo').disabled=!history.length||busy;$('restart').disabled=busy;$('size').disabled=busy;
  $('result').hidden=!ended;if(ended){$('black-score').textContent=score.black;$('white-score').textContent=score.white;}
}
async function start(n) {
  if(busy)return;busy=true;$('notice').textContent='';
  try{game=await api(`/api/new?size=${n}`);history=[];score=null;$('size').value=game.size;}
  catch(e){$('notice').textContent=e.message;if(game)$('size').value=game.size;}
  finally{busy=false;if(game)render();}
}
async function move(action,point=0) {
  if(busy||!game)return;busy=true;$('notice').textContent='';
  try{const next=await api('/api/move',{game,action,point});history.push({game,score});game=next.game;score=next.score;}
  catch(e){$('notice').textContent=e.message;}
  finally{busy=false;render();}
}
function restart(n){if(!game||!history.length||confirm('Start a new game? Current progress will be lost.'))start(n);else $('size').value=game.size;}
$('pass').addEventListener('click',()=>move('pass'));$('undo').addEventListener('click',()=>{if(busy||!history.length)return;({game,score}=history.pop());$('notice').textContent='';render();});$('restart').addEventListener('click',()=>restart(game?.size||19));$('size').addEventListener('change',e=>restart(Number(e.target.value)));
$('board').addEventListener('keydown',e=>{const p=Number(e.target.dataset.point);if(!Number.isInteger(p)||!game)return;const n=game.size,d={ArrowLeft:p%n?-1:0,ArrowRight:p%n<n-1?1:0,ArrowUp:p>=n?-n:0,ArrowDown:p<n*(n-1)?n:0}[e.key];if(d!==undefined){e.preventDefault();$('board').querySelector(`[data-point="${p+d}"]`)?.focus();}});
start(19);
