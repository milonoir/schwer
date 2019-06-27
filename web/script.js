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

  // This function submits the memory allocation size update form using AJAX
  // without redirecting to the /mem endpoint.
  $("#mem-size").submit(function(e) {
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

  const colorBlack = "#000000";
  const colorWhite = "#ffffff";

  const colorRed = "#d53600";
  const colorAmber = "#ffbf00";
  const colorGreen = "#60a917";

  const cpuMeterWidth = 50;
  const cpuMeterHeight = 102;
  const cpuGap = 30;

  function drawCPU(data) {
    var canvas = document.getElementById("cpu-monitor");
    if (canvas.getContext) {
      var ctx = canvas.getContext("2d");

      // Clear canvas.
      ctx.fillStyle = colorWhite;
      ctx.fillRect(0, 0, canvas.width, canvas.height);

      // Handle error message.
      if (typeof(data) === "string") {
        ctx.fillStyle = colorBlack;
        ctx.font = "24px Arial";
        ctx.textAlign = "center";
        ctx.fillText(data, canvas.width / 2, canvas.height / 2);
        return;
      }

      // Determine coordinates of the first CPU meter.
      var startX = (canvas.width - (data.length * cpuMeterWidth + (data.length - 1) * cpuGap)) / 2;
      var startY = (canvas.height - cpuMeterHeight) / 2;

      for (i = 0; i < data.length; i++) {
        var x = startX + i * (cpuMeterWidth + cpuGap);

        // Meter border. Less code than ctx.rect().
        ctx.fillStyle = colorBlack;
        ctx.fillRect(x, startY, cpuMeterWidth, cpuMeterHeight);

        // Fill in the entire meter with the color matching the load level.
        if (data[i] >= 90) {
          ctx.fillStyle = colorRed;
        } else if (data[i] >= 70) {
          ctx.fillStyle = colorAmber;
        } else {
          ctx.fillStyle = colorGreen;
        }
        ctx.fillRect(x + 1, startY + 1, cpuMeterWidth - 2, cpuMeterHeight - 2);

        // Paint the 100% - level part white. This gives the impression of the meters
        // filling up from the bottom.
        ctx.fillStyle = colorWhite;
        ctx.fillRect(x + 1, startY + 1, cpuMeterWidth - 2, cpuMeterHeight - 2 - data[i]);

        // Add textual representation of load level.
        ctx.fillStyle = colorBlack;
        ctx.font = "16px Arial";
        ctx.textAlign = "center";
        ctx.textBaseline = "middle";
        ctx.fillText(data[i].toString() + "%", x + cpuMeterWidth / 2, startY + cpuMeterHeight / 2);
      }
    }
  };

  const memMeterWidth = 602;
  const memMeterHeight = 102;

  function drawMem(data) {
    var canvas = document.getElementById("mem-monitor");
    if (canvas.getContext) {
      var ctx = canvas.getContext("2d");

      // Clear canvas.
      ctx.fillStyle = colorWhite;
      ctx.fillRect(0, 0, canvas.width, canvas.height);

      // Handle error message.
      if (typeof(data) === "string") {
        ctx.fillStyle = colorBlack;
        ctx.font = "24px Arial";
        ctx.textAlign = "center";
        ctx.fillText(data, canvas.width / 2, canvas.height / 2);
        return;
      }

      var x = (canvas.width - memMeterWidth) / 2;
      var y = (canvas.height - memMeterHeight) / 2;

      // Meter border. Less code than ctx.rect().
      ctx.fillStyle = colorBlack;
      ctx.fillRect(x, y, memMeterWidth, memMeterHeight);

      // Paint it all white.
      ctx.fillStyle = colorWhite;
      ctx.fillRect(x + 1, y + 1, memMeterWidth - 2, memMeterHeight - 2);

      if (data.usedpct >= 90) {
        ctx.fillStyle = colorRed;
      } else if (data.usedpct >= 70) {
        ctx.fillStyle = colorAmber;
      } else {
        ctx.fillStyle = colorGreen;
      }
      ctx.fillRect(x + 1, y + 1, memMeterWidth * data.usedpct / 100, memMeterHeight - 2);

      // Add textual representation of stats.
      ctx.fillStyle = colorBlack;
      ctx.font = "16px Arial";

      ctx.textAlign = "center";
      ctx.textBaseline = "hanging";
      ctx.fillText("0 MB", x, y + memMeterHeight + 10);
      ctx.fillText(data.total + " MB", x + memMeterWidth, y + memMeterHeight + 10);

      ctx.textBaseline = "middle";
      ctx.fillText(data.usedpct + "%", x + memMeterWidth / 2, y + memMeterHeight / 2);

      ctx.textAlign = "start";
      ctx.fillText(data.used + " MB", x + 10, y + memMeterHeight / 2);

      ctx.textAlign = "end";
      ctx.fillText(data.available + " MB", x + memMeterWidth - 10, y + memMeterHeight / 2);
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
        drawCPU(data);
      },
      error: function() {
        fb.text(errServerDown);
        drawCPU(errServerDown);
      }
    });
  };

  function memPoll() {
    var fb = $("#mem-monitor-fallback");
    $.ajax({
      type: "GET",
      url: "/mem",
      dataType: "json",
      success: function(data) {
        fb.text(JSON.stringify(data));
        drawMem(data);
        // Update max. allocatable memory value in the form.
        $("#mem-size-input").attr({"max": data.available});
      },
      error: function() {
        fb.text(errServerDown);
        drawMem(errServerDown);
      }
    })
  };

  // Poll CPU and memory once per second.
  setInterval(cpuPoll, 1000);
  setInterval(memPoll, 1000);

});
