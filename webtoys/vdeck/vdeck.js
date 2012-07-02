function fill_vcard(data) {
    // Identification
    $("#input-fullname").val(data.FullName);
    $("#input-firstname").val(data.Name.GivenName);
    $("#input-familyname").val(data.Name.FamilyName);
    $("#input-nickname").val(data.NickName);
    $("#input-birthday").val(data.Birthday);

    // Contact
    $("#vcf-address tbody").empty();
    if (data.Address) {
        for(var i = 0; i < data.Address.length; i++) {
            var addr = data.Address[i];
            var row = $("<tr/>")
                .append($("<td/>").text(addr.POBox))
                .append($("<td/>").text(addr.ExtendedAddr))
                .append($("<td/>").text(addr.Street))
                .append($("<td/>").text(addr.Locality))
                .append($("<td/>").text(addr.Region))
                .append($("<td/>").text(addr.PostalCode))
                .append($("<td/>").text(addr.Country));
            $("#vcf-address tbody").append(row);
        }
    }

    $("#vcf-phone tbody").empty();
    if (data.Tel) {
        for(var i = 0; i < data.Tel.length; i++) {
            var row = $("<td/>").text(data.Tel[i].Value).wrap("<tr/>").parent();
            $("#vcf-phone tbody").append(row);
        }
    }

    $("#vcf-email tbody").empty();
    if (data.Email) {
        for(var i = 0; i < data.Email.length; i++) {
            var row = $("<td/>").text(data.Email[i].Value).wrap("<tr/>").parent();
            $("#vcf-email tbody").append(row);
        }
    }

    // Misc
    $("#input-categories").val(data.Categories);
    $("#input-uid").val(data.Uid);
    $("#input-url").val(data.Url);

    $("#vcf-editor").dialog("open");
};

$(document).ready(function() {
    $("#contacts").jqGrid({
        // caption:  "Contacts",
        url:      "/vdeck/all/",
        datatype: "json",
        mtype:    "GET",
        loadonce: true,
        height:   400,
        scroll:   true,
        colNames: [
        "Full name",
        "Family name",
        "First name",
        "Phone number",
        "Email",
        "Filename",
        ],
        colModel: [
    {name: "fullname"},
        {name: "family_name"},
        {name: "first_name"},
        {name: "phone", align: 'right'},
        {name: "email"},
        {name: "filename"},
        ],
        jsonReader: {repeatitems: false},
        onCellSelect: function(rowid, iCol, content, e) {
            if (iCol == 5) {
                $.get("/vdeck/vcf/" + content, function(data) {
                    $("#vcf-dialog .raw-vcard").text(data);
                    $("#vcf-dialog").dialog("open");
                });
            }
        },
        ondblClickRow: function(rowid, iCol, content, e) {
            var data = $("#contacts").getRowData(rowid);
            if (iCol != 5) {
                $.getJSON("/vdeck/json/" + data.filename, fill_vcard);
            }
        },
    });

    $("#vcf-dialog").dialog({
        autoOpen: false,
        height:   400,
        width:    400,
        modal:    true,
        buttons: {
            "Close": function() { $(this).dialog("close"); },
        },
    });

    $("#vcf-editor").tabs().dialog({
        autoOpen: false,
        height:   650,
        width:    650,
        modal:    true,
        buttons: {
            "Close": function() { $(this).dialog("close"); },
        },
    });
});
