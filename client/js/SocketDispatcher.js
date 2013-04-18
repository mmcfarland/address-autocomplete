// Adapted from https://gist.github.com/ismasan/299789
var SocketDispatcher = function(url){
  var conn = new WebSocket(url),
      callbacks = {};

  this.on = function(name, callback){
    callbacks[name] = callbacks[name] || [];
    callbacks[name].push(callback);
    return this;
  };

  this.send = function(name, data){
    var payload = JSON.stringify({Event: name, Data: data});
    conn.send(payload);
    return this;
  };

  var dispatch = function(name, message){
    var chain = callbacks[name];
    if (typeof chain === 'undefined') return;
    for(var i = 0; i < chain.length; i++){
      chain[i]( message );
    }
  };

  conn.onmessage = function(e){
    var json = JSON.parse(e.data);
    dispatch(json.Event, json.Data);
  };

  conn.onopen = function() {
    // Wire in other websocket native events sans 'on'
    ['close', 'open', 'error'].forEach(function(fn) {
      conn['on' + fn.name] = function(){dispatch(fn, null);};
    });
  };

  // Expose the raw websocket connection
  this.ws = conn;
};