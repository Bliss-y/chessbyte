import { useEffect, useState } from 'react';
import { Interface } from 'readline';
import './App.css';


let fenI: string = "RNBQKBNRPPPPPPPP pppppppprnbqkbnr";

let spaces : string = "";
for (let i=0; i < 8*4; i++){
    spaces += " ";    
}
fenI = fenI.split(" ").join(spaces);

/**
function getAllMoves(position: Position, fen: string) {
    const moves: Array<Position> = [];
    function _getAllMoves(position: Position, piece: string, fen: string, moves: Array<Position>) {
        const newPositions = getNextPositions(position, piece, fen);
        for (let i=0; i < newPositions.length;i++){
            const pieceInPosition = getPiece(newPositions[i], fen);
            if (getPieceSide(piece) === getPieceSide(pieceInPosition)) {
                continue;
            }
            moves.push(newPositions[i]);        
            _getAllMoves(newPositions[i], piece, fen, moves);
        }
    }
    return moves;
}
**/

function Piece({piece, color, size, moveable}: {piece: string, color: string, size: string, moveable:boolean}) {
    
    return <div style={
            {
                backgroundColor: color,
                height: size,
                width: size,
                border: "2px red solid",
            }
        }>{piece} </div>
}

function getPieceSide(p:string) {
    if (p === '' ) {
        return -1
        }
    if(p >= 'a' && p <= 'z'){
        return 1;
    }
    return 0;
}

type Position = {
        x: number,
        y: number,
}


type Game = {
    fen: string,
    direction: 0 | 1,
    turn: 0 | 1,
    clicked: boolean,
    clickedPosition: Position | null,
}

function getPiece(position: Position, fen: string) {
    return fen[position.y * 8 + position.x];
}

function positionIsValid(position: Position) {
    return position.x >=0 && position.y >=0 && position.x < 8 && position.y < 8;
}

// ordering

type MovementFunc = (position: Position, i: number) => Position | null;
type PieceMovements = {
    b: Array<MovementFunc>,
    r: Array<MovementFunc>,
    q: Array<MovementFunc>,
    k: Array<MovementFunc>,
    n: Array<MovementFunc>,
    p: Array<MovementFunc>,
    P: Array<MovementFunc>,
}

// left -up, right-up, left down,  right down
const pieceMovements: PieceMovements = {
    b: [(position: Position, i: number)=>{
            return {
                x: position.x - i-1,
                y: position.y + i + 1
                }
        }, (position: Position, i: number)=>{
            return {
                x: position.x + i + 1,
                y: position.y + i + 1
                }
        }, (position: Position, i: number)=>{
            return {
                x: position.x + i+ 1,
                y: position.y - i- 1
                }
            }, (position: Position, i: number)=>{
            return {
                x: position.x - i -  1,
                y: position.y - i - 1
                }
            }],
    n: [(position: Position, i: number)=> {
            return {
                x: position.x  - 2,
                y: position.y + 1
            }
        },(position: Position, i: number)=> {
            return {
                x: position.x  - 1,
                y: position.y + 2
            }
        },(position: Position, i: number)=> {
            return {
                x: position.x  + 2,
                y: position.y + 1
            }
        },(position: Position, i: number)=> {
            return {
                x: position.x  + 1,
                y: position.y + 2
            }
        },(position: Position, i: number)=> {
            return {
                x: position.x  + 2,
                y: position.y - 1
            }
        },(position: Position, i: number)=> {
            return {
                x: position.x  + 1,
                y: position.y - 1
            }
        },(position: Position, i: number)=> {
            return {
                x: position.x  - 1,
                y: position.y - 1
            }
        },(position: Position, i: number)=> {
            return {
                x: position.x  - 2,
                y: position.y - 1
            }
        }],
    r: [(position: Position, i: number)=>{
            return {
                ...position, x: position.x - i
                }
        },(position: Position, i: number)=>{
            return {
                ...position, y: position.y + i
                }
        },(position: Position, i: number)=>{
            return {
                ...position, x: position.x + i
                }
        },(position: Position, i: number)=>{
            return {
                ...position, y: position.y - i
                }
        },
        ],
    k: [
        (position: Position, i: number)=>{
            return {
                ...position, x: position.x-1
            }
            },
        (position: Position, i: number)=>{
            return {
                ...position, y: position.y+1
            }
            },
        (position: Position, i: number)=>{
            return {
                ...position, x: position.x+1
            }
            },
        (position: Position, i: number)=>{
            return {
                ...position, y: position.y-1
            }
            },
    ]
}

function getNextPositions(position: Position, piece:string, fen: string) {
    if (piece !== "p" && piece !== "P") {
        piece = piece.toLowerCase();
    }
    
    switch (piece) {
        case 'P' :  {
            const positions: Array<Position> = [];
            if ( position.y >= 7){
                return positions;
            }
            positions.push({...position, y: position.y + 1});
            if (position.x >0 && getPieceSide(getPiece({y: position.y + 1, x: position.x-1}, fen)) === 0){
                positions.push({...position, x: position.x - 1});
            }
            if (position.x < 7 && getPieceSide(getPiece({y: position.y + 1, x: position.x + 1 }, fen))){
                positions.push({y: position.y + 1, x: position.x+1});
            }
            return positions;
        }
        case 'p' :  {
            const positions: Array<Position> = [];
            if ( position.y >= 7){
                return positions;
            }
            positions.push({...position, y: position.y - 1});
            if (position.x >0 && getPieceSide(getPiece({y: position.y - 1, x: position.x-1}, fen)) === 0){
                positions.push({y : position.y - 1, x: position.x - 1});
            }
            if (position.x < 7 && getPieceSide(getPiece({y: position.y - 1, x: position.x + 1 }, fen))){
                positions.push({y: position.y - 1, x: position.x+1});
            }
            return positions;
        }
        case 'b' : {
            const moves = [];
            
            break;

            }
        default: {
            return [];
        }
    }
}

type gameType = {
    game?: Game
};

const Board = ({game = {fen: fenI, direction: 0, turn : 0, clicked: false, clickedPosition: null} } : gameType)=> {
    const [gameState, setGame] = useState<Game>({...game});
    const handleClicked = (position:Position, piece:string, moveable:boolean)=> {
        if(gameState.clicked) {
            if (gameState.turn !== gameState.direction) {
                
            }
            if (moveable) {
                // move
            }
            else {
                setGame({...gameState, clicked: false, clickedPosition: null })
                }
        } else {
            if( getPieceSide(piece) === gameState.turn) {
                // show move
            }
        }
    }
    useEffect(()=> {
        }, [])
    let pieces = [];
    if (gameState.direction) {
        let x = 0;
        let y = 0;
        for(let p of gameState.fen) {
            pieces.push(<Piece piece={p} color= {((x + (y % 2 === 0 ? 0 : 1)) % 2 === 0) ? "white" : "black"} size= "50px" moveable = {false}/>)
            if (x === 7 ){
                x = 0;
                y++;
            }else {
                x++
            }
        }
    }
    else{
    for(let i = 7; i >= 0; i--) {
        const row = gameState.fen.slice(i * 8, i * 8 + 8);
        let x = 0;
        for(let p of row){
            pieces.push(<Piece piece={p} color= {(x + (i %2 ===0 ? 0:1)) % 2 === 0 ? "white" : "black"} size= "50px" moveable = {false} />)
            x++;
        }
        }
    }

    return (
    <div style={{
        display: "grid",
        grid: "50px / auto auto auto auto auto auto auto auto",
        color: "red"
        }}>
    {
        pieces
    }
    </div>
    )
}


function App() {
  return (
    <div className="App">
      <header className="App-header">
        <Board/>
      </header>
    </div>
  );
}

export default App;
