(function ($) {
    $.fn.philaddress = function (options) {
        var settings = $.extend({
            url: 'bayntonhill.org:8989',
            minToSend: 3,
            maxResults: 10,
            onError: null,
            onClose: null
        }, options),
        $p = this,
        server = new SocketDispatcher(settings.url),
        $r = $('<ul>', {'class': 'philaddress-list'}),
        $suggest = $('<span>', {'class': 'philadress-suggest'}),
        selectedIdx = -1;

        setupList();

        $p.attr('placeholder', "Enter Address");

        $p.bind("keypress.philaddress", function(e) {
            $p.data('o', $p.val());
            // Cheap optimization - small values (ie. "1") are expensive to search for
            if ($p.val().length > settings.minToSend) {
                server.send('partial', $p.val());
            } else {
                $r.empty().hide();
            }
        }).bind("keydown.philaddress", function(e) {
            $r.children().removeClass('selected');
            if (e.which === 40) {
                if (selectedIdx >= $r.children().length) selectedIdx = -1;
                $($r.children()[++selectedIdx]).addClass('selected');
                $p.val($($r.children()[selectedIdx]).text());
            } else if (e.which === 38) {
                if (selectedIdx === 0) selectedIdx = $r.children().length +1;
                $($r.children()[--selectedIdx]).addClass('selected');
                $p.val($($r.children()[selectedIdx]).text());
            }

        });

        server.on('multiple', function(results) {
            $r.empty();
            var suggestions = $.parseJSON(results);
            if (suggestions && suggestions.length > 0) {
                $r.show();
            } else {
                $r.hide();
                return;
            }
            suggestions.forEach(createSingleSuggestion);
        })
        .on('single', jsonParse(createSingleSuggestion))
        .on('close', function(e) {
            if (settings.onClose && typeof settings.onClose === 'function') {
                settings.onClose();
            }
        })
        .on('error', function(){
            if (settings.onClose && typeof settings.onClose === 'function') {
                settings.onClose();
            }
        });

        function createSingleSuggestion(suggest) {
            var $li = $('<li>', {'class': 'philaddress-list-item'});
            $li.text(suggest.Full);
            $r.append($li);
        }

        function jsonParse(fn) {
            return function(result) {
                fn(JSON.parse(result));
            };
        }

        function setupList() {
            $r.css({
                top: $p.position().top + $p.height(),
                left: $p.position().left
            }).appendTo('body');
        }

        return $p;
    };


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

        // Expose the raw websocket connection
        this.ws = conn;

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
    };

}(jQuery));


