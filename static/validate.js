function validate() {
    var input = document.getElementById('subtitlefile');

    var filename = form.subtitlefile.value;
    var ext = filename.substring(filename.lastIndexOf('.') + 1).toLowerCase();
    if (ext != 'srt' && ext != 'vtt')
    {
        alert('Invalid filetype: ' + ext);
        input.value = '';
        return;
    }

    var filesize = Math.round(input.files[0].size/1024); // filesize in kB
    if (filesize > 200)
    {
        alert('Filesize = ' + filesize + 'kB\nToo large!');
        input.value = '';
        return;
    }
}

function checkFields() {
    var subtitle = form.subtitlefile.value;
    if (!subtitle) {
        alert('No subtitle file is selected.');
        return false;
    }

    var seconds = form.seconds.value;
    if (!seconds) {
        alert('No seconds are entered.');
        return false;
    }

    var filename = form.subtitlefile.value;
    var ext = filename.substring(filename.lastIndexOf('.')).toLowerCase();
    var from = form.from.value;
    alert("from = " + from)
    if (ext != from)
    {
        alert("The file's extension " + ext + " does not match the from field's " + from + ".\n" +
        'Either change the from field to ' + ext + ', or choose a file with extension ' + from + '.');
        return false;
    }

    return true;
}