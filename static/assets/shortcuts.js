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
    var el = document.getElementById("js-dashboard");
    if (el) el.click();
  }
  if (key === 'KeyI') {
    console.log("I Invoices");
    var el = document.getElementById("js-invoices");
    if (el) el.click();
  }
  if (key === 'KeyH') {
    console.log("H Hours");
    var el = document.getElementById("js-hours");
    if (el) el.click();
  }
  if (key === 'KeyE') {
    console.log("E Entities");
    var el = document.getElementById("js-entities");
    if (el) el.click();
  }
  if (key === 'Slash') {
    console.log("/ search");
    var el = document.getElementById("js-search");
    if (el) el.focus();
  }

  if (page === "hours" && key === "KeyN") {
    console.log("Hours New");
    var el = document.getElementById("js-new");
    if (el) el.click();
  }
  if (page === "hour-add" && key === "KeyN") {
    console.log("Hour-add New Entry");
    var el = document.getElementById("hour-day");
    if (el) el.focus();
  }
  if (page === "invoices" && key === "KeyN") {
    console.log("Invoices New");
    var el = document.getElementById("js-new");
    if (el) el.click();
  }
  if (page === "invoices" && key === "KeyB") {
    console.log("Invoices Balance");
    var el = document.getElementById("js-balance");
    if (el) el.click();
  }
}, false);
