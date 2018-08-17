<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8">
    <style>
       .color {
           width: 40px;
           height: 40px;
           display: inline-block;
           border: 1px solid black;
       }
    </style>
  </head>
  <body>
    {{range .}}
        <div>
           <h4>{{.Name}}</h4>
           <div>
               {{range splitColors .Colors}}
                   <div class="color" style="background-color:{{.}}" >
                   </div>
               {{end}}
           </div>
        <div>
    {{end}}
  </body>
</html>
