import Chart from 'chart.js';

// This return to an 24 size array.
// example: array[12] means facebook activity frequency
// in 12 o'clock.
function turnToHourFrequency(friendsData) {
  const FriendsHourFrequency = {};

  for (let i = 0; i < 5; i++) {
    friendsData[i].Activities.forEach((activity) => {
      const data = new Array(24).fill(0);

      var d = new Date(activity.Time * 1000);
      data[d.getHours()]++

      FriendsHourFrequency[friendsData[i].Uid] = data;
    });
  }

  return FriendsHourFrequency;
}

function drawLineChart(friendsData) {
  const FriendsHourFrequency = turnToHourFrequency(friendsData);

  var ctx = document.getElementById("myChart");
  console.log(FriendsHourFrequency);

  const data = {
    labels: [
      "00", "01", "02", "03", "04", "05", "06", "07",
      "08", "09", "10", "11", "12", "13", "14", "15",
      "16", "17", "18", "19", "20", "21", "22", "23",
    ],
    datasets: [
      {
        label: "100000032606394",
        fill: false,
        lineTension: 0.1,
        backgroundColor: "rgba(75,192,192,0.4)",
        borderColor: "rgba(75,192,192,1)",
        borderCapStyle: 'butt',
        borderDash: [],
        borderDashOffset: 0.0,
        borderJoinStyle: 'miter',
        pointBorderColor: "rgba(75,192,192,1)",
        pointBackgroundColor: "#fff",
        pointBorderWidth: 1,
        pointHoverRadius: 5,
        pointHoverBackgroundColor: "rgba(75,192,192,1)",
        pointHoverBorderColor: "rgba(220,220,220,1)",
        pointHoverBorderWidth: 2,
        pointRadius: 1,
        pointHitRadius: 10,
        data: FriendsHourFrequency['100000032606394'],
        spanGaps: false,
      }
    ]
  };

  var myLineChart = new Chart(ctx, {
    type: 'line',
    data: data,
  });
}

(function fetchAndDraw() {
  const xmlhttp = new XMLHttpRequest();

  xmlhttp.onreadystatechange = function() {
    if (xmlhttp.readyState == XMLHttpRequest.DONE) {
      if (xmlhttp.status == 200) {
        // Turn facebook activity data to JSON and draw line chart.
        drawLineChart(JSON.parse(xmlhttp.responseText));
      }
      else if (xmlhttp.status == 400) {
        alert('There was an error 400');
      } else {
        alert('something else other than 200 was returned');
      }
    }
  };

  // "/data" provide facebook activity data.
  xmlhttp.open("GET", "/data", true);
  xmlhttp.send();
})();
