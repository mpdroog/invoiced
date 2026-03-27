window.onerror = function(msg, url, lineNo, columnNo, error) {
  handleErr({
    name: "window.onerror",
    message: msg,
    e: error
  });
};

function prettyErr(res) {
  var msg = "";
  for (var key in res.Fields) {
    if (! res.Fields.hasOwnProperty(key)) {
      continue;
    }

    var line = res.Fields[key];
    msg += "Field (" + key + ") invalid=" + line.join(", ") + "\n";
  }

  document.getElementById("js-error-title").innerText = res.Error;
  document.getElementById("js-error-body").value = msg;
  document.getElementById("js-dialog-error").classList.add('modal-open');
  document.getElementById("js-dialog-error").style.display = 'block';
  document.getElementById("js-dialog-backdrop").classList.remove('hidden');
}

function handleErr(e) {
  if (e.response && e.response.status === 401) {
    location.href = '/static/auth.html';
    return;
  }

  var errorData = {
    pageErrorType: e.name,
    pageMessage: e.message,
    pageURL: document.location.href,
    pageStack: e.e && e.e.stack ? e.e.stack : null,
    serverResponse: e.response ? e.response.data : null,
    serverURL: e.request ? e.request.responseURL : null,
  };
  var splash = document.getElementById("js-splash");
  splash && splash.remove();

  if (e.request && !e.response) {
    document.getElementById("js-dialog-conn-error").classList.add('modal-open');
    document.getElementById("js-dialog-conn-error").style.display = 'block';
    document.getElementById("js-dialog-backdrop").classList.remove('hidden');
    return;
  }

  document.getElementById("js-error-title").innerText = "Oops! Something went wrong...";
  document.getElementById("js-error-body").value = JSON.stringify(errorData);
  document.getElementById("js-dialog-error").classList.add('modal-open');
  document.getElementById("js-dialog-error").style.display = 'block';
  document.getElementById("js-dialog-backdrop").classList.remove('hidden');
}

(function() {
  var btns = document.getElementsByClassName("js-modal-hide");
  Array.prototype.forEach.call(btns, function(m) {
    m.onclick = function() {
      document.getElementById("js-dialog-error").classList.remove("modal-open");
      document.getElementById("js-dialog-error").style.display = 'none';
      document.getElementById("js-dialog-conn-error").classList.remove("modal-open");
      document.getElementById("js-dialog-conn-error").style.display = 'none';
      document.getElementById("js-dialog-backdrop").classList.add('hidden');
    };
  });

  window.addEventListener("click", function(e) {
    if (e.target) {
      var node = e.target;
      while(node !== window && node !== null) {
        if (node.nodeName === "A") {
          if (node.href.substr(-1) === "#") {
            console.log("JS mismatch", node);
            e.preventDefault();
          }
          break;
        }
        node = node.parentNode;
      }
    }
  });

  console.log("JS_ONERR = ON");
})();
