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
			{name: "filename", formatter: linkvcf},
		],
		jsonReader: {repeatitems: false},
	});

        function linkvcf(value, options, row) {
                var link = $("<a/>").attr("href", "/vdeck/vcf/" + row.filename);
                link.text(value);
                return link.wrap("<div>").parent().html();
        };
});
