const $=s=>document.querySelector(s);
let events=[],month=new Date(),filter='all',view='grid',demo=false,selected=null;
month=new Date(month.getFullYear(),month.getMonth(),1);
const zone=Intl.DateTimeFormat().resolvedOptions().timeZone;
$('#timezone').textContent='Часовой пояс: '+zone;
const dateOf=e=>e.date?new Date(e.date.length===10?e.date+'T12:00:00':e.date):null;
const sameDay=(a,b)=>a&&b&&a.getFullYear()===b.getFullYear()&&a.getMonth()===b.getMonth()&&a.getDate()===b.getDate();
const format=(d,opts)=>d.toLocaleDateString('ru-RU',opts);
const escape=s=>String(s??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
const timeText=e=>!e.date?'Дата уточняется':e.date.length===10?'Время уточняется':dateOf(e).toLocaleTimeString('ru-RU',{hour:'2-digit',minute:'2-digit'});
function visible(){let q=$('#search').value.toLocaleLowerCase();return events.filter(e=>(filter==='all'||e.status===filter)&&[e.name,e.venue,e.location].join(' ').toLocaleLowerCase().includes(q));}
function render(){
 $('#month').textContent=format(month,{month:'long',year:'numeric'}).replace(' г.','');
 const found=visible(),inMonth=found.filter(e=>{let d=dateOf(e);return d&&d.getFullYear()===month.getFullYear()&&d.getMonth()===month.getMonth()});
 $('#count').textContent=inMonth.length+' турниров в этом месяце'+(demo?' · ДЕМО':'');
 $('#calendar').hidden=view!=='grid';$('#list').hidden=view!=='list';
 $('#gridView').setAttribute('aria-pressed',view==='grid');$('#listView').setAttribute('aria-pressed',view==='list');
 let html=['ПН','ВТ','СР','ЧТ','ПТ','СБ','ВС'].map(x=>'<div class="weekday">'+x+'</div>').join('');
 let start=new Date(month);start.setDate(1-(month.getDay()+6)%7);
 const cells=Math.ceil(((month.getDay()+6)%7+new Date(month.getFullYear(),month.getMonth()+1,0).getDate())/7)*7;
 for(let i=0;i<cells;i++){let d=new Date(start);d.setDate(start.getDate()+i);html+='<div class="day '+(d.getMonth()!==month.getMonth()?'outside ':'')+(sameDay(d,new Date())?'current':'')+'"><span class="day-number">'+d.getDate()+'</span>';for(const e of found.filter(e=>sameDay(dateOf(e),d)))html+='<button class="event '+(e.status==='completed'?'completed':'')+'" data-event="'+escape(e.id)+'"><strong>'+escape(e.name)+'</strong><small>'+escape(timeText(e))+'</small></button>';html+='</div>';}
 $('#calendar').innerHTML=html;
 $('#list').innerHTML=inMonth.map(e=>'<button class="list-row" data-event="'+escape(e.id)+'"><span class="date-box">'+escape(format(dateOf(e),{day:'2-digit',month:'short'}))+'</span><div><strong>'+escape(e.name)+'</strong><p>'+escape([e.venue,e.location,timeText(e)].filter(Boolean).join(' · '))+'</p></div><span class="arrow">↗</span></button>').join('')||'<div class="empty">'+(events.length?'В этом месяце нет загруженных турниров по выбранным условиям.':'Календарь пока пуст. Подключи API-ключ или открой демо.')+'</div>';
 document.querySelectorAll('[data-event]').forEach(b=>b.onclick=()=>openEvent(events.find(e=>e.id===b.dataset.event)));
 const next=events.filter(e=>dateOf(e)&&e.status!=='completed'&&dateOf(e)>=new Date(new Date().setHours(0,0,0,0))).sort((a,b)=>dateOf(a)-dateOf(b))[0];
 $('#next').innerHTML='<span class="eyebrow">'+(demo?'ДЕМО / ПРИМЕР ТУРНИРА':'БЛИЖАЙШИЙ ТУРНИР')+'</span><h2>'+escape(next?next.name:'В ожидании карда')+'</h2><p>'+escape(next?format(dateOf(next),{day:'numeric',month:'long',weekday:'long'})+' · '+timeText(next):'Новые даты появятся после загрузки данных')+'</p>'+(next?'<p>'+escape([next.venue,next.location].filter(Boolean).join(' · '))+'</p>':'');
}
function openEvent(e){selected=e;$('#detailName').textContent=e.name;$('#detailDate').textContent=e.date?format(dateOf(e),{day:'numeric',month:'long',year:'numeric',weekday:'long'})+' · '+timeText(e):'Дата уточняется';$('#detailVenue').textContent=[e.venue,e.location].filter(Boolean).join(' · ')||'Арена пока не указана';$('#detailStatus').textContent=(demo?'Демонстрационные данные · ':'')+(e.status==='completed'?'Турнир завершён':'По расписанию Cito API');$('#export').disabled=!e.date;$('#detail').showModal();loadCard(e);}
$('#close').onclick=()=>$('#detail').close();$('#detail').onclick=e=>{if(e.target===$('#detail'))$('#detail').close()};
$('#prev').onclick=()=>{month.setMonth(month.getMonth()-1);render()};$('#nextMonth').onclick=()=>{month.setMonth(month.getMonth()+1);render()};$('#today').onclick=()=>{month=new Date(new Date().getFullYear(),new Date().getMonth(),1);render()};
$('#gridView').onclick=()=>{view='grid';render()};$('#listView').onclick=()=>{view='list';render()};$('#search').oninput=render;
 document.querySelectorAll('[data-filter]').forEach(b=>b.onclick=()=>{filter=b.dataset.filter;document.querySelectorAll('[data-filter]').forEach(x=>x.classList.toggle('active',x===b));render()});
$('#demo').onclick=()=>{demo=!demo;$('#demo').textContent=demo?'Выйти из демо':'Посмотреть демо';if(!demo){load();return}let y=new Date().getFullYear(),m=new Date().getMonth();month=new Date(y,m,1);events=[7,14,21,28].map((day,i)=>({id:'demo-'+i,name:['UFC Fight Night: пример','UFC: главный кард','UFC Fight Night: пример','UFC: вечер боёв'][i],date:new Date(y,m,day,22).toISOString(),venue:['Demo Arena','Example Center'][i%2],location:'Демонстрационный турнир',status:day<new Date().getDate()?'completed':'upcoming'}));$('#notice').hidden=false;$('#notice').textContent='Деморежим: вымышленные турниры для знакомства с интерфейсом. Это не настоящее расписание UFC.';$('#sync').textContent='Демо · запросы API не расходуются';render()};
function icsEscape(s){return String(s||'').replace(/\\/g,'\\\\').replace(/\r?\n/g,'\\n').replace(/;/g,'\\;').replace(/,/g,'\\,')}
function stamp(d){return d.toISOString().replace(/[-:]/g,'').replace(/\.\d{3}/,'')}
$('#export').onclick=()=>{let e=selected;const lines=['BEGIN:VCALENDAR','VERSION:2.0','PRODID:-//UFCfans//Calendar//RU','BEGIN:VEVENT','UID:'+icsEscape(encodeURIComponent(e.id))+'@ufcfans.local','DTSTAMP:'+stamp(new Date()),e.date.length===10?'DTSTART;VALUE=DATE:'+e.date.replace(/-/g,''):'DTSTART:'+stamp(dateOf(e)),'SUMMARY:'+icsEscape((demo?'[ДЕМО] ':'')+e.name),'LOCATION:'+icsEscape([e.venue,e.location].filter(Boolean).join(', ')),'DESCRIPTION:'+icsEscape(demo?'Вымышленный демонстрационный турнир':'Расписание Cito API. Время может измениться.'),'END:VEVENT','END:VCALENDAR'];const folded=lines.map(line=>{let result='',size=0;for(const char of line){let n=new TextEncoder().encode(char).length;if(size+n>73){result+='\r\n ';size=1}result+=char;size+=n}return result}).join('\r\n')+'\r\n';let url=URL.createObjectURL(new Blob([folded],{type:'text/calendar;charset=utf-8'}));let a=document.createElement('a');a.href=url;a.download='ufc-event.ics';a.click();setTimeout(()=>URL.revokeObjectURL(url),1000)};
async function load(){try{const r=await fetch('/api/events');if(!r.ok)throw Error();const data=await r.json();events=data.events;$('#setup').hidden=data.configured;$('#notice').hidden=!data.message;$('#notice').textContent=data.message;$('#sync').textContent=data.configured?('Кеш: '+(data.updatedAt.startsWith('0001')?'ещё не загружен':new Date(data.updatedAt).toLocaleString('ru-RU'))+' · '+data.requests+'/'+data.budget+' запросов'):'Без ключа · 0 запросов к API';render()}catch{$('#notice').hidden=false;$('#notice').textContent='Сервер недоступен. Обнови страницу, чтобы повторить загрузку.';$('#sync').textContent='Нет соединения';render()}}
render();load();

let cardRequest=0;
async function loadCard(event){
 const token=++cardRequest;
 $('#boutCount').textContent='';$('#bouts').innerHTML='';$('#cardMessage').textContent='Загружаем пары бойцов…';
 if(demo){$('#cardMessage').textContent='Полный кард доступен для реальных турниров после подключения API.';return}
 try {
  const response=await fetch('/api/events/'+encodeURIComponent(event.id)+'/bouts');
  if(!response.ok)throw Error();
  const data=await response.json();if(token!==cardRequest||selected?.id!==event.id)return;
  $('#boutCount').textContent=data.bouts.length+' боёв';
  $('#cardMessage').textContent=[data.message,data.stale&&data.bouts.length?'Показан сохранённый кард.':'',!data.bouts.length&&!data.message?'Пары бойцов пока не опубликованы.':''].filter(Boolean).join(' ');
  renderBouts(data.bouts);
 }catch{if(token===cardRequest)$('#cardMessage').textContent='Не удалось загрузить кард. Закрой и открой турнир, чтобы повторить.'}
}
const sectionNames={'Main Card':'Основной кард','Prelims':'Предварительный кард','Preliminary Card':'Предварительный кард','Early Prelims':'Ранний предварительный кард'};
const divisions={'Flyweight':'Наилегчайший вес','Bantamweight':'Легчайший вес','Featherweight':'Полулёгкий вес','Lightweight':'Лёгкий вес','Welterweight':'Полусредний вес','Middleweight':'Средний вес','Light Heavyweight':'Полутяжёлый вес','Heavyweight':'Тяжёлый вес',"Women's Strawweight":'Женский минимальный вес',"Women's Flyweight":'Женский наилегчайший вес',"Women's Bantamweight":'Женский легчайший вес',"Women's Featherweight":'Женский полулёгкий вес'};
function renderBouts(bouts){
 const ordered=[...bouts].sort((a,b)=>(a.cardSectionOrder??99)-(b.cardSectionOrder??99)||(a.boutOrder??9999)-(b.boutOrder??9999));
 const groups=new Map();for(const b of ordered){const key=b.cardSection||'Бои турнира';if(!groups.has(key))groups.set(key,[]);groups.get(key).push(b)}
 let html='';for(const [section,rows]of groups){html+='<h4 class="card-section">'+escape(sectionNames[section]||section)+'</h4>';
 for(const b of rows){
  const fighters=Array.isArray(b.fighters)?b.fighters:[];
  const red=fighters.find(f=>f.corner==='red')||fighters[0]||{},blue=fighters.find(f=>f.corner==='blue')||fighters.find(f=>f!==red)||{};
  const cancelled=b.isCancelled||b.status==='cancelled';
  const win=f=>!cancelled&&((b.winnerFighterSlug&&b.winnerFighterSlug===f.fighterSlug)||f.outcome==='win');
  const fighter=f=>'<div class="fighter '+(win(f)?'winner':'')+'">'+(f.fighterSlug?'<button class="fighter-link" data-fighter="'+escape(f.fighterSlug)+'" data-name="'+escape(f.fighterName||f.profile?.name||'Боец')+'">'+escape(f.fighterName||f.profile?.name||'Боец')+' ↗</button>':'<strong>Соперник уточняется</strong>')+'<span>'+escape([f.flag,f.country,f.profile?.recordText].filter(Boolean).join(' · '))+'</span>'+(win(f)?'<b class="winner-label">ПОБЕДИТЕЛЬ</b>':'')+'</div>';
  const weight=b.weightClass?.replace(/ Title$/,'')||'';
  const method={'U-DEC':'Единогласное решение','S-DEC':'Раздельное решение','M-DEC':'Решение большинством','SUB':'Сабмишен','KO/TKO':'KO / TKO'}[b.method]||b.method;
  const outcome=fighters.some(f=>f.outcome==='draw')?'Ничья':fighters.some(f=>['nc','no contest','no_contest'].includes(f.outcome))?'Бой без результата':'';
  const result=cancelled?'Бой отменён'+(b.cancellationReason?' · '+b.cancellationReason:''):[outcome,method,b.resultRound?'Раунд '+b.resultRound:'',b.resultTime].filter(Boolean).join(' · ')||(b.status==='completed'?'Результат пока не опубликован':'Бой запланирован');
  html+='<article class="bout '+(cancelled?'cancelled':'')+'"><div class="bout-meta"><span>'+escape(divisions[weight]||weight||'Весовая категория уточняется')+'</span>'+(b.titleBout?'<span class="title-badge">ТИТУЛЬНЫЙ БОЙ</span>':'')+'</div><div class="matchup">'+fighter(red)+'<span class="versus">VS</span>'+fighter(blue)+'</div><div class="bout-result">'+escape(result)+'</div></article>';
 }
 }$('#bouts').innerHTML=html;
 document.querySelectorAll('[data-fighter]').forEach(button=>button.onclick=()=>openFighter(button.dataset.fighter,button.dataset.name));
}

let profileRequest=0;
function closeFighter(){profileRequest++;$('#fighterProfile').close()}
$('#fighterClose').onclick=closeFighter;$('#fighterBack').onclick=closeFighter;
$('#fighterProfile').addEventListener('cancel',()=>{profileRequest++});
$('#fighterProfile').onclick=e=>{if(e.target===$('#fighterProfile'))closeFighter()};
async function openFighter(slug,name){
 const token=++profileRequest;
 $('#fighterContent').innerHTML='<h2 id="fighterName">'+escape(name)+'</h2>';
 $('#fighterMessage').textContent='Загружаем профиль бойца…';$('#fighterProfile').showModal();
 try{
  const response=await fetch('/api/fighters/'+encodeURIComponent(slug));if(!response.ok)throw Error();
  const data=await response.json();if(token!==profileRequest)return;
  $('#fighterMessage').textContent=[data.message,data.stale&&data.fighter?'Показан сохранённый профиль.':''].filter(Boolean).join(' ');
  if(data.fighter)renderFighter(data.fighter);else if(!data.message)$('#fighterMessage').textContent='Профиль пока не опубликован.';
 }catch{if(token===profileRequest)$('#fighterMessage').textContent='Профиль недоступен. Вернись к карду и открой бойца повторно.'}
}
function renderFighter(f){
 const scalar=v=>['string','number'].includes(typeof v)?String(v):'—';
 const number=v=>v!==null&&v!==undefined&&v!==''&&['string','number'].includes(typeof v)&&Number.isFinite(Number(v))?Number(v):null;
 const metric=(v,factor,unit)=>number(v)===null?'—':(number(v)*factor).toLocaleString('ru-RU',{maximumFractionDigits:1})+' '+unit;
 const stat=(v,percent=false)=>number(v)===null?'—':(number(v)*(percent?100:1)).toLocaleString('ru-RU',{maximumFractionDigits:percent?0:2})+(percent?'%':'');
 const item=(label,value)=>'<div class="profile-item"><dt>'+escape(label)+'</dt><dd>'+escape(value)+'</dd></div>';
 let photo='';try{const url=new URL(f.proxiedImageUrl||f.headshotUrl||f.imageUrl);if(url.protocol==='https:'&&['ufc.com','www.ufc.com','api.citoapi.com'].includes(url.hostname))photo=url.href}catch{}
 const rec=f.record||{},stats=f.stats||{};
 const facts=[['Место рождения',scalar(f.placeOfBirth||f.country)],['Возраст',scalar(f.age)],['Рост',metric(f.heightInches,2.54,'см')],['Вес',metric(f.weightLbs,0.45359237,'кг')],['Размах рук',metric(f.reachInches,2.54,'см')],['Стойка',({Orthodox:'Правосторонняя',Southpaw:'Левосторонняя',Switch:'Смена стойки'})[f.stance]||scalar(f.stance)],['Команда',scalar(f.trainsAt)],['Стиль',scalar(f.fightingStyle)]];
 const statsRows=[['Точность ударов',stats.strikingAccuracy,true],['Защита от ударов',stats.sigStrikeDefense,true],['Точность тейкдаунов',stats.takedownAccuracy,true],['Защита от тейкдаунов',stats.takedownDefense,true],['Значимых ударов / мин',stats.sigStrikesLandedPerMin],['Пропущенных ударов / мин',stats.sigStrikesAbsorbedPerMin],['Тейкдаунов / 15 мин',stats.takedownAvgPer15Min],['Попыток сабмишена / 15 мин',stats.submissionAvgPer15Min]];
 $('#fighterContent').innerHTML='<div class="profile-hero">'+(photo?'<img id="fighterPhoto" src="'+escape(photo)+'" alt="'+escape(f.name)+'" referrerpolicy="no-referrer">':'')+'<div><div class="eyebrow">ПРОФИЛЬ БОЙЦА</div><h2 id="fighterName">'+escape(f.name)+'</h2>'+(f.nickname?'<p class="nickname">«'+escape(f.nickname)+'»</p>':'')+'<p>'+escape(divisions[f.division]||f.division||f.weightClass||'Дивизион не указан')+'</p>'+(f.championStatus==='champion'?'<span class="tag">ЧЕМПИОН</span>':'')+'</div></div><div class="record-grid">'+[['Победы',rec.wins??f.recordWins],['Поражения',rec.losses??f.recordLosses],['Ничьи',rec.draws??f.recordDraws]].map(([label,value])=>'<div><strong>'+escape(scalar(value))+'</strong><span>'+label+'</span></div>').join('')+'</div><h3>О бойце</h3><dl class="profile-grid">'+facts.map(([k,v])=>item(k,v)).join('')+'</dl>'+(f.bio&&typeof f.bio==='string'?'<p class="profile-bio">'+escape(f.bio)+'</p>':'')+'<h3>Статистика карьеры</h3><dl class="profile-grid">'+statsRows.map(([k,v,p])=>item(k,stat(v,p))).join('')+'</dl>'+(Array.isArray(f.rankings)&&f.rankings.length?'<h3>Рейтинги</h3><div class="profile-rankings">'+f.rankings.map(r=>'<p><span>'+escape((r.system==='media'?'Медиарейтинг':r.system==='meta'?'Meta':r.system||'Рейтинг')+' · '+(divisions[r.division]||r.division||''))+'</span><b>'+escape(r.isChampion?'Чемпион':r.rankText||r.rank||'—')+'</b></p>').join('')+'</div>':'')+'<p class="card-note">Данные Cito API. Пропуски означают, что показатель не предоставлен. Профиль обновляется не чаще раза в сутки.</p>';
 const img=$('#fighterPhoto');if(img)img.onerror=()=>{img.hidden=true};
}
