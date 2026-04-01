import QtQuick
import QtQuick.Layouts
import QtQuick.Controls
import org.kde.plasma.components as PlasmaComponents
import org.kde.plasma.core as PlasmaCore
import org.kde.plasma.plasmoid 2.0
import org.kde.kirigami as Kirigami

PlasmoidItem {
    id: root
    width: 300
    height: 220
    
    property string apiUrl: "http://localhost:8080"
    property bool isPlaying: false
    property bool isPaused: false
    property string trackTitle: ""
    property string trackArtist: ""
    property string trackThumbnail: ""
    property int volume: 50
    property string repeatMode: "off"
    property bool shuffle: false
    property bool connected: false
    property bool hasError: false
    property string errorMessage: ""
    property bool showSearchResults: false
    property var searchResults: []
    property var playlists: []
    property int currentView: 0 // 0=nowplaying, 1=search, 2=playlists
    
    function get(url, endpoint) {
        var xhr = new XMLHttpRequest();
        xhr.open("GET", url + endpoint, false);
        xhr.setRequestHeader("Content-Type", "application/json");
        try {
            xhr.send();
            if (xhr.status === 200) {
                return JSON.parse(xhr.responseText);
            }
        } catch (e) {
            console.error("API error:", e);
        }
        return null;
    }
    
    function post(url, endpoint, data) {
        var xhr = new XMLHttpRequest();
        xhr.open("POST", url + endpoint, false);
        xhr.setRequestHeader("Content-Type", "application/json");
        try {
            xhr.send(JSON.stringify(data || {}));
            return xhr.status >= 200 && xhr.status < 300;
        } catch (e) {
            console.error("API error:", e);
            return false;
        }
    }
    
    function updatePlayerState() {
        var state = get(apiUrl, "/player/state");
        
        if (state) {
            connected = true;
            hasError = false;
            isPlaying = state.is_playing;
            isPaused = state.is_paused;
            volume = state.volume;
            repeatMode = state.repeat;
            shuffle = state.shuffle;
            
            if (state.current_track) {
                trackTitle = state.current_track.title;
                trackArtist = state.current_track.artist;
                trackThumbnail = state.current_track.thumbnail_url;
            } else {
                trackTitle = "";
                trackArtist = "";
                trackThumbnail = "";
            }
        } else {
            connected = false;
            hasError = true;
            errorMessage = "Cannot connect to API";
        }
    }
    
    function playPause() {
        post(apiUrl, "/player/pause", {});
        updatePlayerState();
    }
    
    function next() {
        post(apiUrl, "/player/next", {});
        updatePlayerState();
    }
    
    function previous() {
        post(apiUrl, "/player/previous", {});
        updatePlayerState();
    }
    
    function setVolume(vol) {
        volume = vol;
        post(apiUrl, "/player/volume", { volume: vol });
    }
    
    function toggleShuffle() {
        post(apiUrl, "/player/shuffle", {});
        updatePlayerState();
    }
    
    function cycleRepeat() {
        var state = get(apiUrl, "/player/state");
        if (!state) return;
        
        var nextRepeat = "off";
        switch (state.repeat) {
            case "off": nextRepeat = "all"; break;
            case "all": nextRepeat = "one"; break;
            case "one": nextRepeat = "off"; break;
        }
        
        post(apiUrl, "/player/repeat", { repeat: nextRepeat });
        updatePlayerState();
    }
    
    function search(query) {
        if (!query || query.trim() === "") {
            showSearchResults = false;
            searchResults = [];
            return;
        }
        var results = get(apiUrl, "/search?q=" + encodeURIComponent(query));
        if (results && results.results) {
            searchResults = results.results;
            showSearchResults = true;
        }
    }
    
    function playTrack(videoId) {
        post(apiUrl, "/player/play", { video_id: videoId });
        updatePlayerState();
        showSearchResults = false;
    }
    
    function loadPlaylists() {
        var result = get(apiUrl, "/playlists");
        if (result && result.playlists) {
            playlists = result.playlists;
        }
    }
    
    function playPlaylist(playlistId) {
        post(apiUrl, "/playlists/" + playlistId + "/play", {});
        updatePlayerState();
    }
    
    Timer {
        id: timer
        interval: 2000
        running: true
        repeat: true
        onTriggered: updatePlayerState()
    }

    Rectangle {
        id: background
        anchors.fill: parent
        color: "#1a1a1a"
        radius: 8
        
        ColumnLayout {
            anchors.fill: parent
            anchors.margins: 8
            spacing: 6
            
            RowLayout {
                Layout.fillWidth: true
                spacing: 6
                
                Rectangle {
                    width: 56
                    height: 56
                    radius: 4
                    color: "#333"
                    Layout.preferredWidth: 56
                    Layout.preferredHeight: 56
                    
                    Image {
                        id: thumbnail
                        source: trackThumbnail || ""
                        fillMode: Image.PreserveAspectCrop
                        anchors.fill: parent
                        visible: trackThumbnail !== ""
                    }
                    
                    Kirigami.Icon {
                        source: "youtube"
                        anchors.centerIn: parent
                        width: 24
                        height: 24
                        visible: trackThumbnail === ""
                    }
                }
                
                ColumnLayout {
                    Layout.fillWidth: true
                    spacing: 0
                    
                    Text {
                        text: trackTitle || "Not Playing"
                        font.bold: true
                        font.pixelSize: 13
                        color: "white"
                        elide: Text.ElideRight
                        maximumLineCount: 1
                    }
                    
                    Text {
                        text: trackArtist || "No track"
                        font.pixelSize: 11
                        color: "#aaa"
                        elide: Text.ElideRight
                        maximumLineCount: 1
                    }
                }
            }
            
            RowLayout {
                Layout.fillWidth: true
                spacing: 4
                Layout.alignment: Qt.AlignHCenter
                
                PlasmaComponents.ToolButton {
                    icon.name: "media-skip-backward"
                    onClicked: previous()
                    enabled: connected
                    width: 32
                    height: 32
                }
                
                PlasmaComponents.ToolButton {
                    id: playBtn
                    icon.name: isPlaying ? "media-pause" : "media-playback-start"
                    onClicked: playPause()
                    enabled: connected
                    width: 36
                    height: 36
                }
                
                PlasmaComponents.ToolButton {
                    icon.name: "media-skip-forward"
                    onClicked: next()
                    enabled: connected
                    width: 32
                    height: 32
                }
                
                PlasmaComponents.ToolButton {
                    icon.name: shuffle ? "media-playlist-shuffle" : "media-playlist-shuffle"
                    onClicked: toggleShuffle()
                    enabled: connected
                    checkable: true
                    checked: shuffle
                    width: 32
                    height: 32
                    opacity: shuffle ? 1.0 : 0.5
                }
                
                PlasmaComponents.ToolButton {
                    icon.name: repeatMode === "off" ? "media-repeat" : (repeatMode === "one" ? "media-repeat-one" : "media-repeat")
                    onClicked: cycleRepeat()
                    enabled: connected
                    width: 32
                    height: 32
                    opacity: repeatMode !== "off" ? 1.0 : 0.5
                }
            }
            
            RowLayout {
                Layout.fillWidth: true
                spacing: 6
                
                PlasmaComponents.Slider {
                    id: volumeSlider
                    Layout.fillWidth: true
                    from: 0
                    to: 100
                    value: volume
                    onValueChanged: {
                        if (Math.abs(value - volume) > 2) {
                            setVolume(value)
                        }
                    }
                }
                
                Text {
                    text: volume + "%"
                    font.pixelSize: 10
                    color: "#aaa"
                    width: 35
                }
            }
            
            RowLayout {
                Layout.fillWidth: true
                spacing: 4
                
                PlasmaComponents.TextField {
                    id: searchField
                    Layout.fillWidth: true
                    placeholderText: "Search..."
                    onAccepted: search(text)
                    height: 28
                    font.pixelSize: 12
                }
            }
            
            RowLayout {
                Layout.fillWidth: true
                spacing: 4
                
                PlasmaComponents.Button {
                    text: "Now Playing"
                    onClicked: { currentView = 0; showSearchResults = false; }
                    checked: currentView === 0
                    checkable: true
                    height: 24
                    font.pixelSize: 10
                    Layout.fillWidth: true
                }
                
                PlasmaComponents.Button {
                    text: "Playlists"
                    onClicked: { currentView = 2; loadPlaylists(); }
                    checked: currentView === 2
                    checkable: true
                    height: 24
                    font.pixelSize: 10
                    Layout.fillWidth: true
                }
            }
            
            Rectangle {
                Layout.fillWidth: true
                Layout.fillHeight: true
                color: "#222"
                radius: 4
                
                ScrollView {
                    anchors.fill: parent
                    visible: showSearchResults && searchResults.length > 0
                    
                    ColumnLayout {
                        width: parent.width
                        spacing: 2
                        
                        Repeater {
                            model: searchResults
                            delegate: MouseArea {
                                width: parent.width
                                height: 32
                                onClicked: {
                                    if (modelData.track) {
                                        playTrack(modelData.track.video_id);
                                    } else if (modelData.playlist) {
                                        playPlaylist(modelData.playlist.id);
                                    }
                                }
                                
                                RowLayout {
                                    anchors.fill: parent
                                    anchors.margins: 4
                                    spacing: 6
                                    
                                    Kirigami.Icon {
                                        source: modelData.track ? "audio-x-generic" : "folder"
                                        width: 20
                                        height: 20
                                    }
                                    
                                    Text {
                                        text: modelData.track ? (modelData.track.title + " - " + modelData.track.artist) : 
                                               (modelData.playlist ? modelData.playlist.title : "")
                                        color: "white"
                                        font.pixelSize: 11
                                        elide: Text.ElideRight
                                        Layout.fillWidth: true
                                    }
                                }
                            }
                        }
                    }
                }
                
                ScrollView {
                    anchors.fill: parent
                    visible: currentView === 2
                    
                    ColumnLayout {
                        width: parent.width
                        spacing: 2
                        
                        Repeater {
                            model: playlists
                            delegate: MouseArea {
                                width: parent.width
                                height: 36
                                onClicked: playPlaylist(modelData.id)
                                
                                RowLayout {
                                    anchors.fill: parent
                                    anchors.margins: 4
                                    spacing: 6
                                    
                                    Kirigami.Icon {
                                        source: "folder"
                                        width: 24
                                        height: 24
                                    }
                                    
                                    ColumnLayout {
                                        Layout.fillWidth: true
                                        
                                        Text {
                                            text: modelData.title
                                            color: "white"
                                            font.pixelSize: 12
                                            font.bold: true
                                            elide: Text.ElideRight
                                        }
                                        
                                        Text {
                                            text: (modelData.track_count || 0) + " tracks"
                                            color: "#888"
                                            font.pixelSize: 10
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}
