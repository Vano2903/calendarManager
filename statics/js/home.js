async function createSheet() {
    //do a get request to /sheets/new
    var response = await fetch('/sheets/new');
    var sheet = await response.json();
    if (sheet.error) {
        alert(sheet.msg);
        return
    }

    document.getElementById("msg").innerHTML = "(sad la documentazione non é ancora stata creata, chiedere,  al momento, a: Vano, Mora, Perrini per aiuti)il tuo foglio è stato creato con successo, l'url é: " + sheet.data.uri;
    document.getElementById("sheetbtn").innerHTML = "Aggiorna Il Calendario";
    document.getElementById("sheetbtn").onclick = updateCalendar;
}

async function updateCalendar() {
    //do a get request to /calendar/update
    var response = await fetch('/calendar/update');
    var sheet = await response.json();
    if (sheet.error) {
        alert(sheet.msg);
        return
    }

    document.getElementById("msg").innerHTML = "Il calendario è stato aggiornato con successo";
}