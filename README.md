# SC players
![sample](./res/sample.jpg)

함께하는 플레이어의 정보를 볼 수 있습니다.



## 목적
무한정 ID를 생성 할 수 있는 배틀넷 특성상 악질 유저를 구분하기가 쉽지 않음

고민 도중 [원순철님의 자료로 추정되는 PDF][1]를 발견하고 [wireshark]로 패킷 분석을 해본 결과 [배틀코드](#배틀코드)와 IP 주소를 얻을 수 있음을 확인

## 작동 방식
```
me <---6112/udp---> starcraft  
          |
    packet capture --> SC players
```

## 실행 예

[실행 영상 보러가기][2]

## 코드 유의사항
**structures 패키지**가 빠져있는데 이 부분은 직접 구현 해야됨

패킷 분석 or 리버싱을 통해 패킷 구조를 알아내는덴 시간이 많이 소요되진 않음

본인은 리버싱 대신 [wireshark]로 분석함(확실한건 리버싱)


---
#### 배틀코드
 - 블리자드에서 사용하는 ID. 배틀넷 ID와 다름

[1]: http://rosaec.snu.ac.kr/meet/file/20120728c.pdf
[2]: https://youtu.be/1UZxAXiRgRM
[Wireshark]: https://www.wireshark.org/