(function ($) {
    $.fn.philaddress = function (options) {
        var settings = $.extend({
            url: 'ws://parcel.phillygrown.org/suggest/',
            minToSend: 3,
            maxResults: 10,
            onError: null,
            onClose: null
        }, options),
        $p = this,
        server = new SocketDispatcher(settings.url),
        $r = $('<ul>', {'class': 'philaddress-list'}),
        selectedIdx = -1;

        setupList();

        $(document).on('click', 'li.philaddress-list-item', function(e) {
            $p.val(e.target.textContent);
            $r.empty().hide();
        });
        $p.attr('placeholder', "Enter Address");

        function query() {
            $p.data('o', $p.val());
            // Cheap optimization - small values (ie. "1") are expensive to search for
            if ($p.val().length > settings.minToSend) {
                server.send('partial', $p.val());
            } else {
                $r.empty().hide();
            }
        }

        $p.bind("keypress.philaddress", function(e) {
            query();
        })
        .bind("keydown.philaddress", function(e) {
            var c = $r.children(),
                l = $r.children().length;

            c.removeClass('selected');
            // Backspace, delete
            if (e.keyCode === 8 || e.keyCode === 46) {
                query();

            // Up/Down
            } else if (e.keyCode === 40) {
                if (selectedIdx === l-1) {
                    $p.val($p.data('o'));
                    selectedIdx = -1;
                    return;
                }
                $(c[++selectedIdx]).addClass('selected');
                $p.val($(c[selectedIdx]).text());
            } else if (e.keyCode === 38) {
                if (selectedIdx === 0) {
                    $p.val($p.data('o'));
                    selectedIdx = l;
                    return;
                }
                $(c[--selectedIdx]).addClass('selected');
                $p.val($(c[selectedIdx]).text());
            } else if (e.keyCode === 27) {
                $p.val($p.data('o'));
                $r.empty().hide();
            } else if (e.keyCode === 13 || e.keyCode === 9) {
                $r.empty().hide();
            }

        });

        server.on('multiple', function(results) {
            $r.empty();
            selectedIdx = -1;
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
            $li.text(suggest.address + " (" + suggest.owner1 + " " + suggest.owner2 + ")");
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


