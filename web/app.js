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
function openEvent(e){selected=e;$('#detailName').textContent=e.name;$('#detailDate').textContent=e.date?format(dateOf(e),{day:'numeric',month:'long',year:'numeric',weekday:'long'})+' · '+timeText(e):'Дата уточняется';$('#detailVenue').textContent=[e.venue,e.location].filter(Boolean).join(' · ')||'Арена пока не указана';$('#detailStatus').textContent=(demo?'Демонстрационные данные · ':'')+(e.status==='completed'?'Турнир завершён':'По расписанию Cito API');$('#export').disabled=!e.date;$('#detail').showModal();}
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
