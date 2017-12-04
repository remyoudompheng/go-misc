package mail

const mailHtml = `<!DOCTYPE html>
<html>
  <head>
    <title>Webmail</title>
    <meta http-equiv="Content-Type" content="text/html;charset=utf-8"/>
    <meta name="referrer" content="no-referrer">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css"
          integrity="sha384-BVYiiSIFeK1dGmJRAkycuHAHRg32OmUcww7on3RYdg4Va+PmSTsz/K68vbdEjh4u"
          crossorigin="anonymous">
    <script src="https://ajax.googleapis.com/ajax/libs/jquery/1.12.4/jquery.min.js"
            crossorigin="anonymous"></script>
    <script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/js/bootstrap.min.js"
            integrity="sha384-Tc5IQib027qvyjSMfHjOMaLkfuWVxZxUPnCJA7l2mCWNIpG9mGCD8wGNIcPD7Txa"
            crossorigin="anonymous"></script>

    <style>
    @media(min-width: 1600px) {
      .container {
        width: 1440px;
      }
    }

    table.table {
      table-layout: fixed;
    }

    #col-from { width: 20em }
    #col-date { width: 10em }

    td[data-column="from"] {
      white-space: nowrap;
      overflow: hidden;
    }

    #mainHeaders, #fullHeaders {
      white-space: pre-wrap;
    }
    #msgBody {
      border-top: thin solid;
      display: block;
      white-space: pre-wrap;
      max-height: 40em;
      overflow: scroll;
    }
    </style>
  </head>
  <body>
    <div class="container">
      <div class="col-md-2">
        <table class="table table-condensed table-striped">
            <thead><tr><td>Folders</td></tr></thead>
            <tbody id="folder-table">
            </tbody>
        </table>
      </div>
      <div class="col-md-10">
        <table class="table table-condensed table-striped">
          <thead><tr>
              <td id="col-from">From</td>
              <td id="col-subject">Subject</td>
              <td id="col-date">Date</td>
          </tr></thead>
          <tbody id="msg-table"></tbody>
        </table>
      </div>
    </div>

    <template id="row-template">
      <tr>
        <td data-column="from"></td>
        <td data-column="subject"></td>
        <td data-column="date"
            data-toggle="tooltip"
            data-placement="bottom"></td>
      </tr>
    </template>

    <!-- <template id="msg-template"> -->
      <div class="modal fade" id="messageModal" tabindex="-1" role="dialog">
        <div class="modal-dialog modal-lg" role="document">
          <div class="modal-content">
            <div class="modal-body">
                <div class="row" id="mainHeaders"></div>
                <div class="row collapse" id="fullHeaders"></div>
                <div class="row" id="msgBody"></div>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-default"
                      data-toggle="collapse" data-target="#fullHeaders">Headers</button>
              <button type="button" class="btn btn-danger" data-dismiss="modal">Close</button>
            </div>
          </div>
        </div>
      </div>
    <!-- </template> -->

    <script>
    function renderFolders(folders) {
        var $tbl = $("#folder-table");
        var rows = folders.map(function(f) {
            var row = $("<tr><td></td></tr>");
            row.children("td")
               .data("folder", f)
               .text(f);
            return row
        });
        $tbl.append(rows);
        $tbl.find("tr > td").click(function(ev) {
            var f = $(this).data("folder");
            $.get("/mailbox/"+f).success(renderFolder);
        });
    }

    function renderFolder(headers) {
      var tpl = $("#row-template")[0].content;
      $("#msg-table").empty();
      rows = headers.map(function(h) {
        var row = $(document.importNode(tpl, true));
        row.find("tr").attr("data-folder", h.Folder)
        row.find("tr").attr("data-index", h.Index)
        row.find("[data-column='from']").text(h.From);
        row.find("[data-column='subject']").text(h.Subject);
        row.find("[data-column='date']")
           .attr("title", h.FullDate)
           .text(h.PrettyDate);
        return row;
      });
      $("#msg-table").append(rows);
      // attach click event
      $("#msg-table").children("tr").each(function() {
        var folder = $(this).attr("data-folder");
        var idx = $(this).attr("data-index");
        $(this).children("td").click(function(ev) {
          $.get("/message/"+folder+"/"+idx).success(renderMessage);
        });
      });
    }

    function renderMessage(message) {
        $("#messageModal").modal('hide');
        var $mhdr = $("#mainHeaders");
        var $fhdr = $("#fullHeaders");
        var $bdy = $("#msgBody");
        $mhdr.empty();
        $mhdr.text(message.MainHeaders.join("\n"));
        $fhdr.empty();
        $fhdr.text(message.OtherHeaders.join("\n"));
        $fhdr.collapse('hide');
        $bdy.empty();
        $bdy.text(message.Body);
        $("#messageModal").modal('show');
    }

    $(document).ready(function() {
        $.get("/mailboxes/").success(renderFolders);
    });
    </script>
  </body>
</html>`
