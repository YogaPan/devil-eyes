import Chart from 'chart.js';

function turnToHourFrequency(friendsData) {
  const FriendsHourFrequency = {};
  // Get current time.
  const now = new Date();

  friendsData.forEach(friendData => {
    // hourFrequency[12] means facebook activity frequency
    // in 12 o'clock.
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
  ctx.canvas.width = window.innerWidth * 0.98;
  ctx.canvas.height = 500;

  console.log(FriendsHourFrequency);

  // Initialize data.
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
    options: {
      maintainAspectRatio: true,
      responsive: false,
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
