import * as $ from 'jquery'
import * as ProgressBar from 'progressbar.js'

const ws = new WebSocket('ws://localhost:9090/echo');
ws.onmessage =  function incoming(data) {
  bar.animate(data.data/100);
  console.log(data.data);
}

// progressbar.js@1.0.0 version is used
// Docs: http://progressbarjs.readthedocs.org/en/1.0.0/

var bar = new ProgressBar.SemiCircle("#container", {
  strokeWidth: 6,
  color: '#FFEA82',
  trailColor: '#eee',
  trailWidth: 1,
  svgStyle: null,
  text: {
    value: '',
    alignToBottom: true
  },
  from: {color: '#FFEA82'},
  to: {color: '#ED6A5A'},
  // Set default step function for all animate calls
  step: (state, bar) => {
    bar.path.setAttribute('stroke', state.color);
    var value = Math.round(bar.value() * 100);
    if (value === 0) {
      bar.setText('');
    } else {
      bar.setText(value + " %");
    }

    bar.text.style.color = state.color;
  }
});
bar.text.style.fontFamily = '"Raleway", Helvetica, sans-serif';
bar.text.style.fontSize = '2rem';

bar.animate(0.3);  // Number from 0.0 to 1.0

let btn = document.getElementById("coolbutton")
btn.addEventListener("click", (e:Event) => {
  bar.animate(0.7);
})
