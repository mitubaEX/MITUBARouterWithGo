var table;


$(document).ready(function(){
    table = $('#searchResultTable').DataTable({
        stateSave: true,
        searching: false,
        columnDefs: [
            { width: 200, targets: 1 }
        ]
    });
});


$(function() {
    //  searchボタンが押された時
    $('body').on('click', 'button[data-btn-type=search]', function(e) {
        console.log('click btn search');
        //  リクエストの下準備

        var formData = new FormData();
        var $file = $('#file');
        var $birthmark = $('#birthmark');
        var $threshold = $('#threshold');
        console.log($file.attr('name'));
        formData.append($file.attr('name'), $file.prop("files")[0]);
        formData.append($birthmark.attr('name'), $birthmark.val());
        formData.append($threshold.attr('name'), $threshold.val());
        console.log($file);
        console.log($birthmark);
        console.log($threshold);
        console.log(formData.get("file"));
        $.ajax({
            //  リクエストの内容
            type: "POST",
            url: "file",
            data: formData,
            enctype: 'json',
            processData: false,
            contentType: false,
        }).done(function( res ){
            console.log( 'SUCCESS', res);
            console.log( res.toString());
            console.log(JSON.parse(res.toString()));

            table.destroy();

            table = $('#searchResultTable').DataTable({
                data: JSON.parse(res.toString()),
                columns: [
                    { data: 'postedClassFile' },
                    { data: 'resultClassFile' },
                    { data: 'jar' },
                    { data: 'groupId' },
                    { data: 'artifactId' },
                    { data: 'version' },
                    { data: 'sim' }
                ],
                searching: false
                // columnDefs: [
                //     { targets: 1, width: 200 }
                // ]
            });
            return true;

            // var tableHeaderRowCount = 1;
            // var table = document.getElementById('searchResultTable');
            // var rowCount = table.rows.length;
            // for (var i = tableHeaderRowCount; i < rowCount; i++) {
            //     table.deleteRow(tableHeaderRowCount);
            // }
            // var items = [];
            // $.each(JSON.parse(res.toString()), function (key, val) {
            //     items.push("<tr>");
            //     items.push("<td id=''"+key+"''>"+val.postedClassFile+"</td>");
            //     items.push("<td id=''"+key+"''>"+val.resultClassFile+"</td>");
            //     items.push("<td id=''"+key+"''>"+val.jar+"</td>");
            //     items.push("<td id=''"+key+"''>"+val.groupId+"</td>");
            //     items.push("<td id=''"+key+"''>"+val.artifactId+"</td>");
            //     items.push("<td id=''"+key+"''>"+val.version+"</td>");
            //     items.push("<td id=''"+key+"''>"+val.sim+"</td>");
            //     items.push("</tr>");
            // });
            // $("<tbody/>", {html: items.join("")}).appendTo("table");
            // localStorage.setItem('searchResult', res.toString());
            // document.getElementById('searchResult').value = res.toString();
        }).fail(function(jqXHR, textStatus, errorThrown){
            console.log( 'ERROR', jqXHR, textStatus, errorThrown);
        });
        return false;
    });


    //  downloadボタンが押された時
    // $('body').on('click', 'button[data-btn-type=download]', function(e) {
    //     console.log('click btn download');
    //     //  リクエストの下準備
    //     var searchResult = localStorage.getItem('searchResult');
    //
    //     $.ajax({
    //         //  リクエストの内容
    //         type: "POST",
    //         url: "download",
    //         data: searchResult,
    //         enctype: 'text/plain',
    //         processData: false,
    //         contentType: false,
    //     }).done(function( res ){
    //         console.log( 'SUCCESS', res);
    //     }).fail(function(jqXHR, textStatus, errorThrown){
    //         console.log( 'ERROR', jqXHR, textStatus, errorThrown);
    //     });
    //     return false;
    // });
});
