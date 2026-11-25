'use strict';
const GoRules = (() => {
  const key = board => board.join('');
  function newGame(size = 19) {
    if (![9,13,19].includes(size)) throw Error('Use size 9, 13, or 19');
    const board = Array(size*size).fill(0);
    return {size, board, turn:1, captures:[0,0,0], passes:0, history:[key(board)], last:-1, komi:7.5};
  }
  function neighbors(p,n) {
    return [p%n>0?p-1:-1,p%n<n-1?p+1:-1,p>=n?p-n:-1,p<n*(n-1)?p+n:-1].filter(q=>q>=0);
  }
  function group(b,p,n) {
    const stones=[p], seen=new Set(stones); let liberty=false;
    for (const s of stones) for (const q of neighbors(s,n)) {
      if (!b[q]) liberty=true;
      if (b[q]===b[p]&&!seen.has(q)) {seen.add(q);stones.push(q);}
    }
    return {stones,liberty};
  }
  function move(g,action,p) {
    if (g.passes>=2) throw Error('Game has ended');
    if (action==='pass') return {...g,passes:g.passes+1,turn:3-g.turn,last:-1};
    if (action!=='play'||!Number.isInteger(p)||p<0||p>=g.board.length||g.board[p]) throw Error('Choose an empty intersection');
    const b=g.board.slice(), captures=g.captures.slice(); b[p]=g.turn;
    for (const q of neighbors(p,g.size)) if (b[q]===3-g.turn) {
      const {stones,liberty}=group(b,q,g.size);
      if (!liberty) for (const s of stones) {b[s]=0;captures[g.turn]++;}
    }
    if (!group(b,p,g.size).liberty) throw Error('Suicide is not allowed');
    const k=key(b);
    if (g.history.includes(k)) throw Error('Superko: this board position has occurred before');
    return {...g,board:b,captures,history:[...g.history,k],turn:3-g.turn,passes:0,last:p};
  }
  function score(g) {
    const totals=[0,0,g.komi], seen=new Set(), b=g.board;
    for (let p=0;p<b.length;p++) {
      if (b[p]) {totals[b[p]]++;continue;}
      if (seen.has(p)) continue;
      const region=[p], borders=new Set(); seen.add(p);
      for (const s of region) for (const q of neighbors(s,g.size)) {
        if (b[q]) borders.add(b[q]);
        else if (!seen.has(q)) {seen.add(q);region.push(q);}
      }
      if (borders.size===1) totals[[...borders][0]]+=region.length;
    }
    return {black:totals[1],white:totals[2]};
  }
  return {newGame,move,score};
})();
if (typeof module !== 'undefined') module.exports=GoRules;
