document.addEventListener("keyup", function(e) {
  var node = e && e.path && e.path.length > 0 ? e.path[0].nodeName : null;
  node = node ? node : e.target.nodeName;
  if (node === "INPUT" || node === "TEXTAREA") {
    return;
  }

  var key = e.code;
  var page = location.hash.substr(1).split("/")[2];

  if (key === 'KeyD') {
    console.log("D Dashboard");
    document.getElementById("js-dashboard").click();
  }
  if (key === 'KeyI') {
    console.log("I Invoices");
    document.getElementById("js-invoices").click();
  }
  if (key === 'KeyH') {
    console.log("H Hours");
    document.getElementById("js-hours").click();
  }
  if (key === 'KeyE') {
    console.log("E Entities");
    document.getElementById("js-entities").click();
  }
  if (key === 'Slash') {
    console.log("/ search");
    document.getElementById("js-search").focus()
  }

  if (page === "hours" && key === "KeyN") {
    console.log("Hours New");
    document.getElementById("js-new").click();
  }
  if (page === "hour-add" && key === "KeyN") {
    console.log("Hour-add New Entry");
    document.getElementById("hour-day").focus();
  }
  if (page === "invoices" && key === "KeyN") {
    console.log("Invoices New");
    document.getElementById("js-new").click();
  }
  if (page === "invoices" && key === "KeyB") {
    console.log("Invoices Balance");
    document.getElementById("js-balance").click();
  }
}, false);
