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
});
