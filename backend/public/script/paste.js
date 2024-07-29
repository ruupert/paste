(function () {
    $(document).bind('keydown', function (e) {
        var paste;
        if (e.keyCode === 83 && e.ctrlKey) {
            e.preventDefault();
            paste = $('#paste').val();
            var xhr = new XMLHttpRequest();
            req = $.ajax({
                url: '/',
                type: "POST",
                data: paste,
                xhr: function () { return xhr; },
                success: function () { location.href = xhr.responseURL; }
            });
        }
    });
    $(document).ready(function () {
        return $('#paste').focus();
    });
}).call(this);
