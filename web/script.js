$(document).ready(function() {

  // This function submits the cpu percentage update form using AJAX
  // without redirecting to the /cpu endpoint.
  $("#cpuPct").submit(function(e) {
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

});
