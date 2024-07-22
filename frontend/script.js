let fetchParticipants = function (event) {
    fetch('/list-participants')
        .then(response => response.json())
        .then(events => {
            console.log(events);
            const container = document.createElement('div');
            container.classList.add('inline-block');
            container.classList.add('min-w-full');
            container.classList.add('py-2');
            container.classList.add('align-middle');
            container.classList.add('sm:px-6');
            container.classList.add('lg:px-8');

            for (const eventName in events) {
                if (events.hasOwnProperty(eventName)) {
                    const participants = events[eventName];

                    const titleContainer = document.createElement('div');
                    titleContainer.classList.add('sm:flex');
                    titleContainer.classList.add('sm:items-center');

                    const title = document.createElement('h1');
                    title.textContent = eventName;
                    title.classList.add('text-base');
                    title.classList.add('font-semibold');
                    title.classList.add('leading-6');
                    title.classList.add('text-gray-900');
                    container.appendChild(title);

                    const description = document.createElement('p');
                    description.classList.add('mt-2');
                    description.classList.add('text-sm');
                    description.classList.add('text-gray-700');
                    description.textContent = 'Alla deltagare i klassen.'
                    container.appendChild(description);


                    const table = document.createElement('table');
                    table.classList.add('min-w-full');
                    table.classList.add('divide-y');
                    table.classList.add('divide-gray-200');
                    table.classList.add('mb-20');


                    const thead = document.createElement('thead');
                    const headerRow = document.createElement('tr');

                    const headers = ['Startnr', 'Förnamn', 'Efternamn', 'Född', 'Förening/Ort'];
                    headers.forEach(headerText => {
                        const th = document.createElement('th');
                        th.classList.add('py-3');
                        th.classList.add('pl-4');
                        th.classList.add('pr-3');
                        th.classList.add('text-left');
                        th.classList.add('text-xs');
                        th.classList.add('font-medium');
                        th.classList.add('uppercase');
                        th.classList.add('tracking-wide');
                        th.classList.add('text-gray-500');
                        th.classList.add('sm:pl-0');
                        th.textContent = headerText;
                        headerRow.appendChild(th);
                    });

                    thead.appendChild(headerRow);
                    table.appendChild(thead);

                    const tbody = document.createElement('tbody');
                    tbody.classList.add('divide-y');
                    tbody.classList.add('divide-gray-200');
                    tbody.classList.add('bg-white');

                    participants.forEach(participant => {
                        const row = document.createElement('tr');

                        const cells = [
                            participant.BibNumber,
                            participant.FirstName,
                            participant.LastName,
                            participant.Birthdate,
                            participant.Club,
                        ];

                        cells.forEach(cellText => {
                            const td = document.createElement('td');
                            td.classList.add('whitespace-nowrap');
                            td.classList.add('py-4');
                            td.classList.add('pl-4');
                            td.classList.add('pr-3');
                            td.classList.add('text-sm');
                            td.classList.add('font-medium');
                            td.classList.add('text-gray-900');
                            td.classList.add('sm:pl-0');
                            td.textContent = cellText;
                            row.appendChild(td);
                        });

                        tbody.appendChild(row);
                    });

                    table.appendChild(tbody);
                    container.appendChild(table);
                }
            }


            document.getElementById('participants-content').innerHTML = '';
            document.getElementById('participants-content').appendChild(container);
        })
        .catch(error => {
            document.getElementById('participants-content').innerText = 'Error listing participants: ' + error;
        });
}


document.getElementById('watch-form').addEventListener('submit', function (event) {
    event.preventDefault();
    const filePath = document.getElementById('filePath').value;
    fetch('/start-watch', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({filePath})
    })
        .then(response => response.text())
        .then(data => {
            document.getElementById('watch-feedback').innerText = data;
        })
        .catch(error => {
            document.getElementById('watch-feedback').innerText = 'Error starting watch: ' + error;
        });
});

document.getElementById('sheets-form').addEventListener('submit', function (event) {
    event.preventDefault();
    const sheetID = document.getElementById('sheetID').value;
    const sheetName = document.getElementById('sheetName').value;
    fetch('/google-sheets', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({sheetID, sheetName})
    })
        .then(response => response.text())
        .then(data => {
            document.getElementById('sheets-feedback').innerText = data;
        })
        .catch(error => {
            document.getElementById('sheets-feedback').innerText = 'Error submitting Google Sheets info: ' + error;
        });
});

document.getElementById('read-startlista').addEventListener('submit', function (event) {
    event.preventDefault();
    const primaryEventName = document.getElementById('eventName').value;
    const participantsSheetName = document.getElementById('participantsSheetName').value;
    fetch('/read-startlista', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({primaryEventName: primaryEventName, participantsSheetName: participantsSheetName})
    })
        .then(response => response.json())
        .then(data => {
            document.getElementById('startlista-content').innerHTML = 'Startlista importerad';

        })
        .catch(error => {
            document.getElementById('startlista-content').innerText = 'Error reading Startlista: ' + error;
        });
});


document.getElementById('list-participants').addEventListener('click', fetchParticipants);

document.addEventListener('alpine:init', () => {
    Alpine.data('appData', () => ({
        tab: 'config',
        participantsSheetName: '',
        eventName: '',
        sheetID: '',
        sheetName: '',
        filePath: '',

        init() {
            this.$watch('tab', () => {
                localStorage.setItem('tab', this.tab);
            });
            this.$watch('participantsSheetName', () => {
                localStorage.setItem('participantsSheetName', this.participantsSheetName);
            });
            this.$watch('eventName', () => {
                localStorage.setItem('eventName', this.eventName);
            });
            this.$watch('sheetID', () => {
                localStorage.setItem('sheetID', this.sheetID);
            });
            this.$watch('sheetName', () => {
                localStorage.setItem('sheetName', this.sheetName);
            });
            this.$watch('filePath', () => {
                localStorage.setItem('filePath', this.filePath);
            });
        },

        reset() {
            this.tab = 'config';
            this.participantsSheetName = '';
            this.eventName = '';
            this.sheetID = '';
            this.sheetName = '';
            this.filePath = '';
        },

        loadData() {
            // load from localStorage
            const tab = localStorage.getItem('tab');
            const participantsSheetName = localStorage.getItem('participantsSheetName');
            const eventName = localStorage.getItem('eventName');
            const sheetID = localStorage.getItem('sheetID');
            const sheetName = localStorage.getItem('sheetName');
            const filePath = localStorage.getItem('filePath');
            if (tab) {
                this.tab = tab;

                if (tab === 'participants') {
                    fetchParticipants();
                }
            }
            if (participantsSheetName) {
                this.participantsSheetName = participantsSheetName;
            }
            if (eventName) {
                this.eventName = eventName;
            }
            if (sheetID) {
                this.sheetID = sheetID;
            }
            if (sheetName) {
                this.sheetName = sheetName;
            }
            if (filePath) {
                this.filePath = filePath;
            }
        },
    }));
});


