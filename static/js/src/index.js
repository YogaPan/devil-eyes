import Chart from 'chart.js';

/*
 * Get params from url.
 *
 * From stackoverflow:
 * http://stackoverflow.com/questions/901115/how-can-i-get-query-string-values-in-javascript
 *
 * usage:
 * query string: ?foo=lorem&bar=&baz
 * var foo = getParameterByName('foo'); => "lorem"
 * var bar = getParameterByName('bar'); => "" (present with empty value)
 * var baz = getParameterByName('baz'); => "" (present with no value)
 * var qux = getParameterByName('qux'); => null (absent)
 */
function getParameterByName(name, url) {
  if (!url) url = window.location.href;
  name = name.replace(/[\[\]]/g, "\\$&");
  var regex = new RegExp("[?&]" + name + "(=([^&#]*)|&|#|$)"),
  results = regex.exec(url);
  if (!results) return null;
  if (!results[2]) return '';
  return decodeURIComponent(results[2].replace(/\+/g, " "));
}

function turnToHourFrequency(friendsData) {
  const FriendsHourFrequency = {};
  // Get current time.
  const now = new Date();

  friendsData.forEach(friendData => {
    // example:
    // hourFrequency[12] = 3
    // means this guy active 3 times at 12 o'clock.
    const hourFrequency = new Array(24).fill(0);

    friendData.Activities.forEach(activity => {
      const activityTime = new Date(activity.Time * 1000);

      // Only need today's data.
      if (activityTime.getDate() === now.getDate()) {
        hourFrequency[activityTime.getHours()]++
      }
    });

    FriendsHourFrequency[friendData.Uid] = hourFrequency;
  });

  return FriendsHourFrequency;
}

function drawLineChart(friendsData) {
  const FriendsHourFrequency = turnToHourFrequency(friendsData);

  const ctx = document.getElementById("myChart").getContext('2d');
  // multi 0.98 to prevent horizontal scrollbar.
  ctx.canvas.width = window.innerWidth * 0.98;
  ctx.canvas.height = 500;

  // Debug data.
  // console.log(FriendsHourFrequency);

  // Initialize lineChartData.
  const lineChartData = {};

  // Add 'labels' elements to object (x axis).
  // This means 24 hour.
  lineChartData.labels = [
    "00", "01", "02", "03", "04", "05", "06", "07",
    "08", "09", "10", "11", "12", "13", "14", "15",
    "16", "17", "18", "19", "20", "21", "22", "23",
  ];
  lineChartData.datasets = [];

  // Only show 5 friends in chart.
  let count = 0;
  for (var uid in FriendsHourFrequency) {
    lineChartData.datasets.push({});

    const dataset = lineChartData.datasets[count];

    dataset.data = FriendsHourFrequency[uid];
    dataset.label = uid;
    dataset.fill = false;
    dataset.tension = 0.2;
    dataset.backgroundColor = "rgba(75,192,192,0.4)";
    dataset.borderColor = "rgba(75,192,192,1)";

    count++;
    // Only show 5 friends in chart.
    if (count >= 5) {
      break;
    }
  }

  const myLineChart = new Chart(ctx, {
    type: 'line',
    data: lineChartData,
    options: { // These options let you can adjust canvas size.
      maintainAspectRatio: true,
      responsive: true,
    },
  });
}

(function fetchAndDraw() {
  const xmlhttp = new XMLHttpRequest();

  xmlhttp.onreadystatechange = function() {
    if (xmlhttp.readyState == XMLHttpRequest.DONE) {
      if (xmlhttp.status == 200) {
        // Turn facebook activity data into JSON and draw line chart.
        drawLineChart(JSON.parse(xmlhttp.responseText));
      }
      else if (xmlhttp.status == 400) {
        console.log('There was an error 400');
      } else {
        console.log('something else other than 200 was returned');
      }
    }
  };

  // Path "/data" provide all friends facebook activity data.
  // Path "/data/1003123" only provide facebook id 1003123's data.
  if (getParameterByName('uid') !== null) {
    xmlhttp.open("GET", `/data/${getParameterByName('uid')}`, true);
  } else {
    xmlhttp.open("GET", '/data', true);
  }
  xmlhttp.send();
})();
