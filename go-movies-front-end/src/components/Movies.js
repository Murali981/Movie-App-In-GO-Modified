import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

const Movies = () => {
  const [movies, setMovies] = useState([]);

  ///// We will use  useEffect hook to call a fake backend API.....

  useEffect(() => {
    const headers = new Headers();
    headers.append("Content-Type", "application/json");

    const requestOptions = {
      method: "GET",
      headers: headers,
    };

    fetch(`http://localhost:8080/movies`, requestOptions).then((response) =>
      response
        .json()
        .then((data) => {
          setMovies(data);
        })
        .catch((err) => {
          console.log(err);
        })
    );

    // let moviesList = [
    //   {
    //     id: 1,
    //     title: "Highlander",
    //     release_date: "1986-03-07",
    //     runtime: 116,
    //     mpaa_rating: "A",
    //     description: "some long description",
    //   },
    //   {
    //     id: 2,
    //     title: "Raiders of the last arc",
    //     release_date: "1981-06-12",
    //     runtime: 115,
    //     mpaa_rating: "PG-13",
    //     description: "some long description",
    //   },
    // ];
    // setMovies(moviesList);
  }, []);

  return (
    <div>
      <h2>Movies</h2>
      <hr />
      <table className="table table-striped table-hover">
        <thead>
          <tr>
            <th>Movie</th>
            <th>Release Date</th>
            <th>Rating</th>
          </tr>
        </thead>
        <tbody>
          {movies.map((m) => (
            <tr key={m.id}>
              <td>
                <Link to={`/movies/${m.id}`}>{m.title}</Link>
              </td>
              <td>{m.release_date}</td>
              <td>{m.mpaa_rating}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

export default Movies;
