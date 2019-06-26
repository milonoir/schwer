$(document).ready(function() {

  // This function submits the cpu percentage update form using AJAX
  // without redirecting to the /cpu endpoint.
  $("#cpu-pct").submit(function(e) {
    e.preventDefault();

    var form = $(this);
    var url = form.attr("action");

    $.ajax({
      type: "POST",
      url: url,
      data: form.serialize(),
      success: function(data) {
        alert(data);
      }
    });
  });

  const meterWidth = 50;
  const meterHeight = 102;
  const gap = 30;

  function draw(data) {
    var canvas = document.getElementById("cpu-monitor");
    if (canvas.getContext) {
      var ctx = canvas.getContext("2d");

      // Clear canvas.
      ctx.fillStyle = "#ffffff";
      ctx.fillRect(0, 0, canvas.width, canvas.height);

      // Handle error message.
      if (typeof(data) === "string") {
        ctx.fillStyle = "#000000";
        ctx.font = "24px Arial";
        ctx.textAlign = "center";
        ctx.fillText(data, canvas.width / 2, canvas.height / 2);
        return;
      }

      // Determine coordinates of the first CPU meter.
      var startX = (canvas.width - (data.length * meterWidth + (data.length - 1) * gap)) / 2;
      var startY = (canvas.height - meterHeight) / 2;

      for (i = 0; i < data.length; i++) {
        var x = startX + i * (meterWidth + gap);

        // Meter border. Less code than ctx.rect().
        ctx.fillStyle = "#000000";
        ctx.fillRect(x, startY, meterWidth, meterHeight);

        // Fill in the entire meter with the color matching the load level.
        if (data[i] >= 90) {
          ctx.fillStyle = "#d53600"; // Red above 90%.
        } else if (data[i] >= 70) {
          ctx.fillStyle = "#ffbf00"; // Amber above 70%.
        } else {
          ctx.fillStyle = "#60a917"; // Green below 70%.
        }
        ctx.fillRect(x + 1, startY + 1, meterWidth - 2, meterHeight - 2);

        // Paint the 100% - level part white.
        ctx.fillStyle = "#ffffff";
        ctx.fillRect(x + 1, startY + 1, meterWidth - 2, meterHeight - 2 - data[i]);

        // Add textual representation of load level. This gives the impression of the meters
        // filling up from the bottom.
        ctx.fillStyle = "#000000";
        ctx.font = "16px Arial";
        ctx.textAlign = "center";
        ctx.fillText(data[i].toString(), x + meterWidth / 2, startY + meterHeight / 2);
      }
    }
  };

  const errServerDown = "Unable to connect to Schwer server.";

  // This functions updates CPU utilisation levels.
  function cpuPoll() {
    var fb = $("#cpu-monitor-fallback");
    $.ajax({
      type: "GET",
      url: "/cpu",
      dataType: "json",
      success: function(data) {
        fb.text(data);
        draw(data);
      },
      error: function() {
        fb.text(errServerDown);
        draw(errServerDown);
      }
    });
  };

  // Execute cpuPoll() once per second.
  setInterval(cpuPoll, 1000);

});
