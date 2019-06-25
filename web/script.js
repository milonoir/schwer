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

  // This functions updates CPU utilisation levels.
  function cpuPoll() {
    var status = $("#cpu-monitor");
    $.ajax({
      type: "GET",
      url: "/cpu",
      dataType: "json",
      success: function(data) {
        status.text(data);
      }
    });
  };

  // Execute cpuPoll() once per second.
  setInterval(cpuPoll, 1000);

});
